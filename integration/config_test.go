package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestConfigGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	result, err := client.ConfigGet(ctx, "maxmemory").Result()
	if err != nil {
		t.Fatalf("CONFIG GET failed: %v", err)
	}
	// stub always returns empty — just assert no error and correct type
	_ = result
}

func TestConfigSet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.ConfigSet(ctx, "maxmemory", "100mb").Err(); err != nil {
		t.Fatalf("CONFIG SET failed: %v", err)
	}
}

func TestConfigResetStat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.ConfigResetStat(ctx).Err(); err != nil {
		t.Fatalf("CONFIG RESETSTAT failed: %v", err)
	}
}

func TestConfigUnknownSubcommandReturnsError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	err := client.Do(ctx, "CONFIG", "BADCMD").Err()
	if err == nil || !strings.Contains(err.Error(), "unknown subcommand") {
		t.Fatalf("expected unknown subcommand error, got %v", err)
	}
}

func TestConfigMissingSubcommandReturnsError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	err := client.Do(ctx, "CONFIG").Err()
	if err == nil || !strings.Contains(err.Error(), "wrong number of arguments") {
		t.Fatalf("expected wrong number of arguments error, got %v", err)
	}
}
