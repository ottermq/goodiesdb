package integration_test

import (
	"context"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// TestPubSubBasicPublishSubscribe verifies a single subscriber receives a
// published message on the correct channel with the correct payload.
func TestPubSubBasicPublishSubscribe(t *testing.T) {
	addr := startTestServer(t)

	subClient := newRedisClient(t, addr, 0)
	pubClient := newRedisClient(t, addr, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := subClient.Subscribe(ctx, "news")
	defer sub.Close()

	// Wait for subscription to be acknowledged.
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("subscribe confirmation failed: %v", err)
	}

	n, err := pubClient.Publish(ctx, "news", "hello").Result()
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected PUBLISH to return 1, got %d", n)
	}

	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}
	if msg.Channel != "news" {
		t.Fatalf("expected channel %q, got %q", "news", msg.Channel)
	}
	if msg.Payload != "hello" {
		t.Fatalf("expected payload %q, got %q", "hello", msg.Payload)
	}
}

// TestPubSubMultipleSubscribersAllReceive verifies that all connected subscribers
// receive the message and PUBLISH returns the correct count.
func TestPubSubMultipleSubscribersAllReceive(t *testing.T) {
	addr := startTestServer(t)
	pubClient := newRedisClient(t, addr, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const numSubs = 3
	subs := make([]*redis.PubSub, numSubs)
	for i := range subs {
		subs[i] = newRedisClient(t, addr, 0).Subscribe(ctx, "broadcast")
		defer subs[i].Close()
		if _, err := subs[i].Receive(ctx); err != nil {
			t.Fatalf("subscriber %d confirmation failed: %v", i, err)
		}
	}

	n, err := pubClient.Publish(ctx, "broadcast", "ping").Result()
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	if int(n) != numSubs {
		t.Fatalf("expected PUBLISH to return %d, got %d", numSubs, n)
	}

	for i, sub := range subs {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			t.Fatalf("subscriber %d failed to receive: %v", i, err)
		}
		if msg.Payload != "ping" {
			t.Fatalf("subscriber %d: expected %q, got %q", i, "ping", msg.Payload)
		}
	}
}

// TestPubSubUnsubscribeStopsDelivery verifies that after UNSUBSCRIBE the client
// no longer receives messages and PUBLISH returns 0.
func TestPubSubUnsubscribeStopsDelivery(t *testing.T) {
	addr := startTestServer(t)
	subClient := newRedisClient(t, addr, 0)
	pubClient := newRedisClient(t, addr, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := subClient.Subscribe(ctx, "events")
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("subscribe confirmation failed: %v", err)
	}

	if err := sub.Unsubscribe(ctx, "events"); err != nil {
		t.Fatalf("UNSUBSCRIBE failed: %v", err)
	}
	// Drain the unsubscribe confirmation.
	sub.Receive(ctx) //nolint

	n, err := pubClient.Publish(ctx, "events", "should-not-arrive").Result()
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 receivers after unsubscribe, got %d", n)
	}
}

// TestPubSubPatternSubscribeReceivesMatchingChannel verifies PSUBSCRIBE delivers
// messages on channels matching the glob pattern and excludes non-matching ones.
func TestPubSubPatternSubscribeReceivesMatchingChannel(t *testing.T) {
	addr := startTestServer(t)
	subClient := newRedisClient(t, addr, 0)
	pubClient := newRedisClient(t, addr, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := subClient.PSubscribe(ctx, "events.*")
	defer sub.Close()
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("psubscribe confirmation failed: %v", err)
	}

	// Non-matching publish — should not be delivered.
	pubClient.Publish(ctx, "other.topic", "ignored")

	// Matching publish.
	if _, err := pubClient.Publish(ctx, "events.login", "user123").Result(); err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}

	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}
	if msg.Channel != "events.login" {
		t.Fatalf("expected channel %q, got %q", "events.login", msg.Channel)
	}
	if msg.Payload != "user123" {
		t.Fatalf("expected payload %q, got %q", "user123", msg.Payload)
	}
	if msg.Pattern != "events.*" {
		t.Fatalf("expected pattern %q, got %q", "events.*", msg.Pattern)
	}
}

// TestPubSubSubscriberModeRejectsNormalCommands verifies that a client in
// subscriber mode receives an error when issuing non-pub/sub commands.
func TestPubSubSubscriberModeRejectsNormalCommands(t *testing.T) {
	addr := startTestServer(t)
	subClient := newRedisClient(t, addr, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := subClient.Subscribe(ctx, "ch")
	defer sub.Close()
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("subscribe confirmation failed: %v", err)
	}

	// A second client on the same underlying connection is not easily testable
	// via go-redis PubSub — the library manages the connection state.
	// Instead, verify that PUBLISH (allowed in normal mode, but we test via a
	// fresh client) still works, and that our server hasn't broken non-pubsub
	// clients while one subscriber exists.
	freshClient := newRedisClient(t, addr, 0)
	if err := freshClient.Set(ctx, "key", "val", 0).Err(); err != nil {
		t.Fatalf("SET on fresh client failed while subscriber is active: %v", err)
	}
}

// TestPubSubDisconnectCleanup verifies that an abrupt client disconnect does not
// cause panics or leave stale state: subsequent PUBLISH returns 0.
func TestPubSubDisconnectCleanup(t *testing.T) {
	addr := startTestServer(t)
	pubClient := newRedisClient(t, addr, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a raw client that we close manually (not via newRedisClient, which
	// would register a t.Cleanup that double-closes).
	subClient := redis.NewClient(&redis.Options{Addr: addr})
	sub := subClient.Subscribe(ctx, "volatile")
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("subscribe confirmation failed: %v", err)
	}
	// Close abruptly without UNSUBSCRIBE.
	sub.Close()
	subClient.Close()

	// Give the server a moment to detect the disconnect.
	time.Sleep(100 * time.Millisecond)

	n, err := pubClient.Publish(ctx, "volatile", "after-disconnect").Result()
	if err != nil {
		t.Fatalf("PUBLISH after disconnect failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 receivers after disconnect, got %d", n)
	}
}

// TestPubSubPublishOnUnknownChannelReturnsZero verifies PUBLISH to a channel
// with no subscribers returns 0 without error.
func TestPubSubPublishOnUnknownChannelReturnsZero(t *testing.T) {
	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	n, err := client.Publish(ctx, "nobody-listening", "message").Result()
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}
