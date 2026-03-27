package integration_test

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/andrelcunha/goodiesdb/internal/core/server"
)

func TestPingAndSetGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("PING failed: %v", err)
	}
	if pong != "PONG" {
		t.Fatalf("expected PONG, got %q", pong)
	}

	if err := client.Set(ctx, "hello", "world", 0).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	got, err := client.Get(ctx, "hello").Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if got != "world" {
		t.Fatalf("expected %q, got %q", "world", got)
	}
}

func TestMissingKeyReturnsRedisNil(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	_, err := client.Get(ctx, "missing").Result()
	if err != redis.Nil {
		t.Fatalf("expected redis.Nil, got %v", err)
	}
}

func TestDatabaseSelectionIsolatedByClientDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	db0 := newRedisClient(t, addr, 0)
	db1 := newRedisClient(t, addr, 1)

	if err := db0.Set(ctx, "shared-key", "db0-value", 0).Err(); err != nil {
		t.Fatalf("SET on db0 failed: %v", err)
	}

	if err := db1.Set(ctx, "shared-key", "db1-value", 0).Err(); err != nil {
		t.Fatalf("SET on db1 failed: %v", err)
	}

	got0, err := db0.Get(ctx, "shared-key").Result()
	if err != nil {
		t.Fatalf("GET on db0 failed: %v", err)
	}
	got1, err := db1.Get(ctx, "shared-key").Result()
	if err != nil {
		t.Fatalf("GET on db1 failed: %v", err)
	}

	if got0 != "db0-value" {
		t.Fatalf("expected db0 value %q, got %q", "db0-value", got0)
	}
	if got1 != "db1-value" {
		t.Fatalf("expected db1 value %q, got %q", "db1-value", got1)
	}
}

func TestExpireMakesKeyDisappear(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "temp-key", "value", 0).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	if err := client.Expire(ctx, "temp-key", time.Second).Err(); err != nil {
		t.Fatalf("EXPIRE failed: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	_, err := client.Get(ctx, "temp-key").Result()
	if err != redis.Nil {
		t.Fatalf("expected redis.Nil after expiration, got %v", err)
	}
}

func TestIncrAndDecr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	value, err := client.Incr(ctx, "counter").Result()
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected first INCR result to be 1, got %d", value)
	}

	value, err = client.Incr(ctx, "counter").Result()
	if err != nil {
		t.Fatalf("second INCR failed: %v", err)
	}
	if value != 2 {
		t.Fatalf("expected second INCR result to be 2, got %d", value)
	}

	value, err = client.Decr(ctx, "counter").Result()
	if err != nil {
		t.Fatalf("DECR failed: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected DECR result to be 1, got %d", value)
	}
}

func TestListCommands(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	length, err := client.LPush(ctx, "letters", "a", "b", "c").Result()
	if err != nil {
		t.Fatalf("LPUSH failed: %v", err)
	}
	if length != 3 {
		t.Fatalf("expected LPUSH length 3, got %d", length)
	}

	values, err := client.LRange(ctx, "letters", 0, -1).Result()
	if err != nil {
		t.Fatalf("LRANGE failed: %v", err)
	}
	expected := []string{"c", "b", "a"}
	if strings.Join(values, ",") != strings.Join(expected, ",") {
		t.Fatalf("expected LRANGE result %v, got %v", expected, values)
	}

	popped, err := client.LPop(ctx, "letters").Result()
	if err != nil {
		t.Fatalf("LPOP failed: %v", err)
	}
	if popped != "c" {
		t.Fatalf("expected LPOP result %q, got %q", "c", popped)
	}
}

func TestCommandErrorsVisibleThroughRedisClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "not-an-int", "hello", 0).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	_, err := client.Do(ctx, "SET", "only-key").Result()
	if err == nil || !strings.Contains(err.Error(), "wrong number of arguments") {
		t.Fatalf("expected wrong number of arguments error, got %v", err)
	}

	_, err = client.Incr(ctx, "not-an-int").Result()
	if err == nil || !strings.Contains(err.Error(), "not an integer") {
		t.Fatalf("expected integer error, got %v", err)
	}
}

func TestDelRemovesExistingKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "delete-me", "value", 0).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	deleted, err := client.Del(ctx, "delete-me").Result()
	if err != nil {
		t.Fatalf("DEL failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected DEL to return 1, got %d", deleted)
	}

	_, err = client.Get(ctx, "delete-me").Result()
	if err != redis.Nil {
		t.Fatalf("expected redis.Nil after DEL, got %v", err)
	}
}

func TestListEdgeCommands(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.RPush(ctx, "numbers", "one", "two", "three", "four").Err(); err != nil {
		t.Fatalf("RPUSH failed: %v", err)
	}

	popped, err := client.Do(ctx, "RPOP", "numbers", 2).StringSlice()
	if err != nil {
		t.Fatalf("RPOP count failed: %v", err)
	}
	expectedPopped := []string{"three", "four"}
	if strings.Join(popped, ",") != strings.Join(expectedPopped, ",") {
		t.Fatalf("expected RPOP count result %v, got %v", expectedPopped, popped)
	}

	if err := client.LTrim(ctx, "numbers", 0, 0).Err(); err != nil {
		t.Fatalf("LTRIM failed: %v", err)
	}

	values, err := client.LRange(ctx, "numbers", 0, -1).Result()
	if err != nil {
		t.Fatalf("LRANGE after LTRIM failed: %v", err)
	}
	expectedRemaining := []string{"one"}
	if strings.Join(values, ",") != strings.Join(expectedRemaining, ",") {
		t.Fatalf("expected trimmed list %v, got %v", expectedRemaining, values)
	}
}

func TestStringAndIntrospectionCommands(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "greeting", "hello world", 0).Err(); err != nil {
		t.Fatalf("SET greeting failed: %v", err)
	}
	if err := client.RPush(ctx, "animals", "cat", "dog").Err(); err != nil {
		t.Fatalf("RPUSH animals failed: %v", err)
	}

	length, err := client.StrLen(ctx, "greeting").Result()
	if err != nil {
		t.Fatalf("STRLEN failed: %v", err)
	}
	if length != 11 {
		t.Fatalf("expected STRLEN 11, got %d", length)
	}

	substr, err := client.GetRange(ctx, "greeting", 0, 4).Result()
	if err != nil {
		t.Fatalf("GETRANGE failed: %v", err)
	}
	if substr != "hello" {
		t.Fatalf("expected GETRANGE result %q, got %q", "hello", substr)
	}

	valueType, err := client.Type(ctx, "animals").Result()
	if err != nil {
		t.Fatalf("TYPE failed: %v", err)
	}
	if valueType != "list" {
		t.Fatalf("expected TYPE result %q, got %q", "list", valueType)
	}

	keys, err := client.Keys(ctx, "*g*").Result()
	if err != nil {
		t.Fatalf("KEYS failed: %v", err)
	}
	sort.Strings(keys)
	expectedKeys := []string{"greeting"}
	sort.Strings(expectedKeys)
	if strings.Join(keys, ",") != strings.Join(expectedKeys, ",") {
		t.Fatalf("expected KEYS result %v, got %v", expectedKeys, keys)
	}
}

func startTestServer(t *testing.T) string {
	t.Helper()

	cfg := server.NewConfig()
	cfg.Host = "127.0.0.1"
	cfg.Port = "0"
	cfg.UseAOF = false
	cfg.UseRDB = false
	cfg.DataDir = t.TempDir()
	cfg.Version = "test"

	srv := server.NewServer(cfg)
	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start()
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if addr := srv.Addr(); addr != "" {
			return registerServerCleanup(t, srv, errCh, addr)
		}

		select {
		case err := <-errCh:
			t.Fatalf("server failed to start: %v", err)
		default:
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("server did not become ready within %s", 2*time.Second)
	return ""
}

func registerServerCleanup(t *testing.T, srv *server.Server, errCh <-chan error, addr string) string {
	t.Helper()

	t.Cleanup(func() {
		srv.Shutdown()
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("server shutdown returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for server shutdown")
		}
	})

	return addr
}

func newRedisClient(t *testing.T, addr string, db int) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("failed to close redis client: %v", err)
		}
	})

	return client
}
