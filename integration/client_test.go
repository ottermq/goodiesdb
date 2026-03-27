package integration_test

import (
	"context"
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
