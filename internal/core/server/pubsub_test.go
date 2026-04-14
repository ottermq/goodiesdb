package server

import (
	"net"
	"testing"
	"time"
)

// mockConn returns a pair of connected net.Conn for testing.
func mockConn() (net.Conn, net.Conn) {
	c1, c2 := net.Pipe()
	return c1, c2
}

func TestSubscribeAndReceive(t *testing.T) {
	b := newPubSubBroker()
	c, _ := mockConn()
	defer c.Close()

	count := b.Subscribe(c, "news")
	if count != 1 {
		t.Fatalf("expected 1 subscription, got %d", count)
	}

	ch := b.GetConnChan(c)
	n := b.Publish("news", "hello")
	if n != 1 {
		t.Fatalf("expected 1 receiver, got %d", n)
	}

	select {
	case msg := <-ch:
		expected := encodeMessage("news", "hello")
		if string(msg) != string(expected) {
			t.Fatalf("unexpected message:\ngot  %q\nwant %q", msg, expected)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestMultipleSubscribersReceive(t *testing.T) {
	b := newPubSubBroker()
	c1, _ := mockConn()
	c2, _ := mockConn()
	c3, _ := mockConn()
	defer c1.Close()
	defer c2.Close()
	defer c3.Close()

	b.Subscribe(c1, "news")
	b.Subscribe(c2, "news")
	b.Subscribe(c3, "news")

	ch1 := b.GetConnChan(c1)
	ch2 := b.GetConnChan(c2)
	ch3 := b.GetConnChan(c3)

	n := b.Publish("news", "broadcast")
	if n != 3 {
		t.Fatalf("expected 3 receivers, got %d", n)
	}

	for i, ch := range []<-chan []byte{ch1, ch2, ch3} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatalf("client %d timed out waiting for message", i+1)
		}
	}
}

func TestPublishReturnsZeroWithNoSubscribers(t *testing.T) {
	b := newPubSubBroker()
	n := b.Publish("empty", "msg")
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	b := newPubSubBroker()
	c, _ := mockConn()
	defer c.Close()

	b.Subscribe(c, "news")
	remaining := b.Unsubscribe(c, "news")
	if remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", remaining)
	}

	n := b.Publish("news", "after-unsub")
	if n != 0 {
		t.Fatalf("expected 0 receivers after unsubscribe, got %d", n)
	}
}

func TestUnsubscribeAllCleansUp(t *testing.T) {
	b := newPubSubBroker()
	c, _ := mockConn()
	defer c.Close()

	b.Subscribe(c, "ch1")
	b.Subscribe(c, "ch2")
	b.PSubscribe(c, "events.*")

	b.UnsubscribeAll(c)

	if n := b.Publish("ch1", "x"); n != 0 {
		t.Fatalf("expected 0 after UnsubscribeAll, got %d on ch1", n)
	}
	if n := b.Publish("events.login", "x"); n != 0 {
		t.Fatalf("expected 0 after UnsubscribeAll, got %d on events.login", n)
	}
	if b.GetConnChan(c) != nil {
		t.Fatal("expected delivery channel to be removed after UnsubscribeAll")
	}
}

func TestPatternSubscribeReceivesMatchingChannel(t *testing.T) {
	b := newPubSubBroker()
	c, _ := mockConn()
	defer c.Close()

	count := b.PSubscribe(c, "events.*")
	if count != 1 {
		t.Fatalf("expected 1 subscription, got %d", count)
	}

	ch := b.GetConnChan(c)
	n := b.Publish("events.login", "user123")
	if n != 1 {
		t.Fatalf("expected 1 receiver, got %d", n)
	}

	select {
	case msg := <-ch:
		expected := encodePMessage("events.*", "events.login", "user123")
		if string(msg) != string(expected) {
			t.Fatalf("unexpected message:\ngot  %q\nwant %q", msg, expected)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for pattern message")
	}
}

func TestPatternSubscribeDoesNotReceiveNonMatching(t *testing.T) {
	b := newPubSubBroker()
	c, _ := mockConn()
	defer c.Close()

	b.PSubscribe(c, "events.*")
	ch := b.GetConnChan(c)

	n := b.Publish("other.topic", "data")
	if n != 0 {
		t.Fatalf("expected 0 receivers for non-matching channel, got %d", n)
	}

	select {
	case msg := <-ch:
		t.Fatalf("received unexpected message: %q", msg)
	case <-time.After(50 * time.Millisecond):
		// correct: nothing delivered
	}
}

func TestSlowConsumerDoesNotBlockPublisher(t *testing.T) {
	b := newPubSubBroker()
	c, _ := mockConn()
	defer c.Close()

	b.Subscribe(c, "flood")

	// Fill the buffer and then some — publisher must not block.
	done := make(chan struct{})
	go func() {
		for i := 0; i < pubsubDeliveryBufSize+10; i++ {
			b.Publish("flood", "msg")
		}
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on slow consumer")
	}
}

func TestPublishCountsExactAndPatternTogether(t *testing.T) {
	b := newPubSubBroker()
	cExact, _ := mockConn()
	cPattern, _ := mockConn()
	defer cExact.Close()
	defer cPattern.Close()

	b.Subscribe(cExact, "events.login")
	b.PSubscribe(cPattern, "events.*")

	n := b.Publish("events.login", "payload")
	if n != 2 {
		t.Fatalf("expected 2 receivers (1 exact + 1 pattern), got %d", n)
	}
}

func TestSingleConnReceivesBothExactAndPattern(t *testing.T) {
	b := newPubSubBroker()
	c, _ := mockConn()
	defer c.Close()

	b.Subscribe(c, "events.login")
	b.PSubscribe(c, "events.*")

	ch := b.GetConnChan(c)

	// Publish counts as 1 (one conn), but delivers 2 messages (exact + pattern).
	n := b.Publish("events.login", "data")
	if n != 1 {
		t.Fatalf("expected 1 counted receiver for single conn, got %d", n)
	}

	msgs := 0
	for msgs < 2 {
		select {
		case <-ch:
			msgs++
		case <-time.After(time.Second):
			t.Fatalf("timed out after %d messages, expected 2", msgs)
		}
	}
}
