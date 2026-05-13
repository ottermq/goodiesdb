package integration_test

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/ottermq/goodiesdb/internal/core/server"
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

func TestRPushAppendsValuesInOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	length, err := client.RPush(ctx, "queue", "a", "b", "c").Result()
	if err != nil {
		t.Fatalf("RPUSH failed: %v", err)
	}
	if length != 3 {
		t.Fatalf("expected RPUSH length 3, got %d", length)
	}

	values, err := client.LRange(ctx, "queue", 0, -1).Result()
	if err != nil {
		t.Fatalf("LRANGE after RPUSH failed: %v", err)
	}
	expected := []string{"a", "b", "c"}
	if strings.Join(values, ",") != strings.Join(expected, ",") {
		t.Fatalf("expected RPUSH order %v, got %v", expected, values)
	}
}

func TestLPopWithCountReturnsMultipleItems(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.RPush(ctx, "lpop:count", "one", "two", "three").Err(); err != nil {
		t.Fatalf("RPUSH setup failed: %v", err)
	}

	values, err := client.Do(ctx, "LPOP", "lpop:count", 2).StringSlice()
	if err != nil {
		t.Fatalf("LPOP count failed: %v", err)
	}

	expected := []string{"one", "two"}
	if strings.Join(values, ",") != strings.Join(expected, ",") {
		t.Fatalf("expected LPOP count result %v, got %v", expected, values)
	}
}

func TestRenameRenamesExistingKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "old-key", "value", 0).Err(); err != nil {
		t.Fatalf("SET old-key failed: %v", err)
	}
	if err := client.Rename(ctx, "old-key", "new-key").Err(); err != nil {
		t.Fatalf("RENAME failed: %v", err)
	}

	value, err := client.Get(ctx, "new-key").Result()
	if err != nil {
		t.Fatalf("GET new-key failed: %v", err)
	}
	if value != "value" {
		t.Fatalf("expected renamed value %q, got %q", "value", value)
	}

	_, err = client.Get(ctx, "old-key").Result()
	if err != redis.Nil {
		t.Fatalf("expected old-key to be missing after RENAME, got %v", err)
	}
}

func TestEchoAndInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	echo, err := client.Echo(ctx, "hello world").Result()
	if err != nil {
		t.Fatalf("ECHO failed: %v", err)
	}
	if echo != "hello world" {
		t.Fatalf("expected ECHO result %q, got %q", "hello world", echo)
	}

	info, err := client.Info(ctx).Result()
	if err != nil {
		t.Fatalf("INFO failed: %v", err)
	}
	if !strings.Contains(info, "# Server") {
		t.Fatalf("expected INFO output to contain server section, got %q", info)
	}
}

func TestFlushDbAndFlushAll(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	db0 := newRedisClient(t, addr, 0)
	db1 := newRedisClient(t, addr, 1)

	if err := db0.Set(ctx, "flushdb:key", "db0", 0).Err(); err != nil {
		t.Fatalf("SET flushdb:key failed: %v", err)
	}
	if err := db1.Set(ctx, "flushall:key", "db1", 0).Err(); err != nil {
		t.Fatalf("SET flushall:key failed: %v", err)
	}

	if err := db0.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("FLUSHDB failed: %v", err)
	}

	_, err := db0.Get(ctx, "flushdb:key").Result()
	if err != redis.Nil {
		t.Fatalf("expected db0 key to be missing after FLUSHDB, got %v", err)
	}
	value, err := db1.Get(ctx, "flushall:key").Result()
	if err != nil {
		t.Fatalf("expected db1 key to survive FLUSHDB, got %v", err)
	}
	if value != "db1" {
		t.Fatalf("expected db1 key value %q, got %q", "db1", value)
	}

	if err := db0.FlushAll(ctx).Err(); err != nil {
		t.Fatalf("FLUSHALL failed: %v", err)
	}

	_, err = db1.Get(ctx, "flushall:key").Result()
	if err != redis.Nil {
		t.Fatalf("expected db1 key to be missing after FLUSHALL, got %v", err)
	}
}

func TestScanFindsMatchingKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "scan:1", "a", 0).Err(); err != nil {
		t.Fatalf("SET scan:1 failed: %v", err)
	}
	if err := client.Set(ctx, "scan:2", "b", 0).Err(); err != nil {
		t.Fatalf("SET scan:2 failed: %v", err)
	}
	if err := client.Set(ctx, "other", "c", 0).Err(); err != nil {
		t.Fatalf("SET other failed: %v", err)
	}

	keys, cursor, err := client.Scan(ctx, 0, "scan:*", 10).Result()
	if err != nil {
		t.Fatalf("SCAN failed: %v", err)
	}
	if cursor != 0 {
		t.Fatalf("expected SCAN cursor to be 0, got %d", cursor)
	}
	sort.Strings(keys)
	expected := []string{"scan:1", "scan:2"}
	if strings.Join(keys, ",") != strings.Join(expected, ",") {
		t.Fatalf("expected SCAN keys %v, got %v", expected, keys)
	}
}

func TestAuthAcceptsValidPassword(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	result, err := client.Do(ctx, "AUTH", "guest").Text()
	if err != nil {
		t.Fatalf("AUTH failed: %v", err)
	}
	if result != "OK" {
		t.Fatalf("expected AUTH result %q, got %q", "OK", result)
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

func TestExistsCountsMatchingKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "exists:one", "1", 0).Err(); err != nil {
		t.Fatalf("SET exists:one failed: %v", err)
	}
	if err := client.Set(ctx, "exists:two", "2", 0).Err(); err != nil {
		t.Fatalf("SET exists:two failed: %v", err)
	}

	count, err := client.Exists(ctx, "exists:one").Result()
	if err != nil {
		t.Fatalf("EXISTS single key failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected EXISTS single key result 1, got %d", count)
	}

	count, err = client.Exists(ctx, "exists:one", "exists:two", "exists:missing").Result()
	if err != nil {
		t.Fatalf("EXISTS multiple keys failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected EXISTS multiple keys result 2, got %d", count)
	}
}

func TestExpireReportsExistingAndMissingKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "expiring-key", "value", 0).Err(); err != nil {
		t.Fatalf("SET expiring-key failed: %v", err)
	}

	ok, err := client.Expire(ctx, "expiring-key", time.Second).Result()
	if err != nil {
		t.Fatalf("EXPIRE existing key failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected EXPIRE to return true for an existing key")
	}

	ok, err = client.Expire(ctx, "missing-key", time.Second).Result()
	if err != nil {
		t.Fatalf("EXPIRE missing key failed: %v", err)
	}
	if ok {
		t.Fatalf("expected EXPIRE to return false for a missing key")
	}
}

func TestTTLReportsExpiringPersistentAndMissingKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "ttl:expiring", "value", 0).Err(); err != nil {
		t.Fatalf("SET ttl:expiring failed: %v", err)
	}
	if err := client.Set(ctx, "ttl:persistent", "value", 0).Err(); err != nil {
		t.Fatalf("SET ttl:persistent failed: %v", err)
	}
	if err := client.Expire(ctx, "ttl:expiring", 5*time.Second).Err(); err != nil {
		t.Fatalf("EXPIRE ttl:expiring failed: %v", err)
	}

	ttl, err := client.TTL(ctx, "ttl:expiring").Result()
	if err != nil {
		t.Fatalf("TTL expiring key failed: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("expected TTL for expiring key to be positive, got %v", ttl)
	}

	ttl, err = client.TTL(ctx, "ttl:persistent").Result()
	if err != nil {
		t.Fatalf("TTL persistent key failed: %v", err)
	}
	if ttl != -1 {
		t.Fatalf("expected TTL for persistent key to be -1, got %v", ttl)
	}

	ttl, err = client.TTL(ctx, "ttl:missing").Result()
	if err != nil {
		t.Fatalf("TTL missing key failed: %v", err)
	}
	if ttl != -2 {
		t.Fatalf("expected TTL for missing key to be -2, got %v", ttl)
	}
}

func TestSetNXOnlySetsMissingKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	created, err := client.SetNX(ctx, "setnx:key", "first", 0).Result()
	if err != nil {
		t.Fatalf("SETNX first call failed: %v", err)
	}
	if !created {
		t.Fatalf("expected first SETNX call to create the key")
	}

	created, err = client.SetNX(ctx, "setnx:key", "second", 0).Result()
	if err != nil {
		t.Fatalf("SETNX second call failed: %v", err)
	}
	if created {
		t.Fatalf("expected second SETNX call to leave the key unchanged")
	}

	value, err := client.Get(ctx, "setnx:key").Result()
	if err != nil {
		t.Fatalf("GET after SETNX failed: %v", err)
	}
	if value != "first" {
		t.Fatalf("expected SETNX key to keep original value %q, got %q", "first", value)
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

func TestHashCommands(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	added, err := client.HSet(ctx, "session:1", "user_id", "1", "username", "andre", "avatar_color", "#3498DB").Result()
	if err != nil {
		t.Fatalf("initial HSET failed: %v", err)
	}
	if added != 3 {
		t.Fatalf("expected initial HSET to add 3 fields, got %d", added)
	}

	added, err = client.HSet(ctx, "session:1", "username", "andre-updated", "avatar_url", "https://example.com/avatar.png").Result()
	if err != nil {
		t.Fatalf("second HSET failed: %v", err)
	}
	if added != 1 {
		t.Fatalf("expected second HSET to add 1 new field, got %d", added)
	}

	username, err := client.HGet(ctx, "session:1", "username").Result()
	if err != nil {
		t.Fatalf("HGET existing field failed: %v", err)
	}
	if username != "andre-updated" {
		t.Fatalf("expected HGET username %q, got %q", "andre-updated", username)
	}

	_, err = client.HGet(ctx, "session:1", "missing").Result()
	if err != redis.Nil {
		t.Fatalf("expected redis.Nil for missing hash field, got %v", err)
	}

	all, err := client.HGetAll(ctx, "session:1").Result()
	if err != nil {
		t.Fatalf("HGETALL failed: %v", err)
	}
	expectedAll := map[string]string{
		"user_id":      "1",
		"username":     "andre-updated",
		"avatar_color": "#3498DB",
		"avatar_url":   "https://example.com/avatar.png",
	}
	if len(all) != len(expectedAll) {
		t.Fatalf("expected HGETALL size %d, got %d", len(expectedAll), len(all))
	}
	for field, expected := range expectedAll {
		if got := all[field]; got != expected {
			t.Fatalf("expected HGETALL[%q] = %q, got %q", field, expected, got)
		}
	}

	values, err := client.HMGet(ctx, "session:1", "user_id", "missing", "avatar_url").Result()
	if err != nil {
		t.Fatalf("HMGET failed: %v", err)
	}
	if len(values) != 3 {
		t.Fatalf("expected HMGET to return 3 values, got %d", len(values))
	}
	if values[0] != "1" {
		t.Fatalf("expected HMGET first value %q, got %v", "1", values[0])
	}
	if values[1] != nil {
		t.Fatalf("expected HMGET missing field to be nil, got %v", values[1])
	}
	if values[2] != "https://example.com/avatar.png" {
		t.Fatalf("expected HMGET third value to match avatar_url, got %v", values[2])
	}

	exists, err := client.HExists(ctx, "session:1", "avatar_color").Result()
	if err != nil {
		t.Fatalf("HEXISTS failed: %v", err)
	}
	if !exists {
		t.Fatalf("expected HEXISTS to return true for avatar_color")
	}

	length, err := client.HLen(ctx, "session:1").Result()
	if err != nil {
		t.Fatalf("HLEN failed: %v", err)
	}
	if length != 4 {
		t.Fatalf("expected HLEN 4, got %d", length)
	}

	keys, err := client.HKeys(ctx, "session:1").Result()
	if err != nil {
		t.Fatalf("HKEYS failed: %v", err)
	}
	sort.Strings(keys)
	expectedKeys := []string{"avatar_color", "avatar_url", "user_id", "username"}
	if strings.Join(keys, ",") != strings.Join(expectedKeys, ",") {
		t.Fatalf("expected HKEYS %v, got %v", expectedKeys, keys)
	}

	vals, err := client.HVals(ctx, "session:1").Result()
	if err != nil {
		t.Fatalf("HVALS failed: %v", err)
	}
	sort.Strings(vals)
	expectedVals := []string{"#3498DB", "1", "andre-updated", "https://example.com/avatar.png"}
	sort.Strings(expectedVals)
	if strings.Join(vals, ",") != strings.Join(expectedVals, ",") {
		t.Fatalf("expected HVALS %v, got %v", expectedVals, vals)
	}

	deleted, err := client.HDel(ctx, "session:1", "avatar_color", "missing").Result()
	if err != nil {
		t.Fatalf("HDEL failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected HDEL to delete 1 field, got %d", deleted)
	}

	length, err = client.HLen(ctx, "session:1").Result()
	if err != nil {
		t.Fatalf("HLEN after HDEL failed: %v", err)
	}
	if length != 3 {
		t.Fatalf("expected HLEN after HDEL to be 3, got %d", length)
	}
}

func TestHashCommandsRespectExpiration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.HSet(ctx, "hash:expiring", "field", "value").Err(); err != nil {
		t.Fatalf("HSET failed: %v", err)
	}
	if err := client.Expire(ctx, "hash:expiring", time.Second).Err(); err != nil {
		t.Fatalf("EXPIRE failed: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	_, err := client.HGet(ctx, "hash:expiring", "field").Result()
	if err != redis.Nil {
		t.Fatalf("expected redis.Nil for expired hash field, got %v", err)
	}

	length, err := client.HLen(ctx, "hash:expiring").Result()
	if err != nil {
		t.Fatalf("HLEN on expired hash failed: %v", err)
	}
	if length != 0 {
		t.Fatalf("expected HLEN on expired hash to be 0, got %d", length)
	}
}

func TestHashCommandsReturnWrongTypeForNonHashKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "plain:string", "value", 0).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	_, err := client.HGet(ctx, "plain:string", "field").Result()
	if err == nil || !strings.Contains(err.Error(), "WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE from HGET on string key, got %v", err)
	}

	_, err = client.HSet(ctx, "plain:string", "field", "value").Result()
	if err == nil || !strings.Contains(err.Error(), "WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE from HSET on string key, got %v", err)
	}
}

func TestAOFRestartPreservesValuesWithSpaces(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := server.NewConfig()
	cfg.Host = "127.0.0.1"
	cfg.Port = "0"
	cfg.UseAOF = true
	cfg.UseRDB = false
	cfg.DataDir = t.TempDir()
	cfg.Version = "test"

	srv, errCh, addr := startServerWithConfig(t, cfg)
	client := newRedisClient(t, addr, 0)

	if err := client.Set(ctx, "greeting", "hello world", 0).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	if err := client.RPush(ctx, "people", "Andre Cunha", "Maria Clara").Err(); err != nil {
		t.Fatalf("RPUSH failed: %v", err)
	}
	if err := client.HSet(ctx, "profile", "display_name", "Andre Cunha", "bio", "{\"summary\":\"hello world\"}").Err(); err != nil {
		t.Fatalf("HSET failed: %v", err)
	}

	stopServer(t, srv, errCh)

	srv, errCh, addr = startServerWithConfig(t, cfg)
	client = newRedisClient(t, addr, 0)

	greeting, err := client.Get(ctx, "greeting").Result()
	if err != nil {
		t.Fatalf("GET after restart failed: %v", err)
	}
	if greeting != "hello world" {
		t.Fatalf("expected greeting to survive restart, got %q", greeting)
	}

	people, err := client.LRange(ctx, "people", 0, -1).Result()
	if err != nil {
		t.Fatalf("LRANGE after restart failed: %v", err)
	}
	expectedPeople := []string{"Andre Cunha", "Maria Clara"}
	if strings.Join(people, ",") != strings.Join(expectedPeople, ",") {
		t.Fatalf("expected people list %v after restart, got %v", expectedPeople, people)
	}

	displayName, err := client.HGet(ctx, "profile", "display_name").Result()
	if err != nil {
		t.Fatalf("HGET after restart failed: %v", err)
	}
	if displayName != "Andre Cunha" {
		t.Fatalf("expected hash field to survive restart, got %q", displayName)
	}

	stopServer(t, srv, errCh)
}

func TestClientSetNameAndGetName(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)

	// Use a pinned connection — ClientSetName is per-connection state.
	conn := newRedisClient(t, addr, 0).Conn()
	defer conn.Close()

	if err := conn.ClientSetName(ctx, "myconn").Err(); err != nil {
		t.Fatalf("CLIENT SETNAME failed: %v", err)
	}
	name, err := conn.ClientGetName(ctx).Result()
	if err != nil {
		t.Fatalf("CLIENT GETNAME after set failed: %v", err)
	}
	if name != "myconn" {
		t.Fatalf("expected CLIENT GETNAME %q, got %q", "myconn", name)
	}

	fresh := newRedisClient(t, addr, 0).Conn()
	defer fresh.Close()
	_, err = fresh.ClientGetName(ctx).Result()
	if err != redis.Nil {
		t.Fatalf("expected redis.Nil for CLIENT GETNAME on unnamed connection, got %v", err)
	}

	err = conn.ClientSetName(ctx, "bad name").Err()
	if err == nil {
		t.Fatalf("expected CLIENT SETNAME with spaces to return an error")
	}
}

func TestClientID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	c1 := newRedisClient(t, addr, 0)
	c2 := newRedisClient(t, addr, 0)

	id1, err := c1.ClientID(ctx).Result()
	if err != nil {
		t.Fatalf("CLIENT ID on c1 failed: %v", err)
	}
	if id1 <= 0 {
		t.Fatalf("expected CLIENT ID to be positive, got %d", id1)
	}

	id2, err := c2.ClientID(ctx).Result()
	if err != nil {
		t.Fatalf("CLIENT ID on c2 failed: %v", err)
	}
	if id2 <= 0 {
		t.Fatalf("expected CLIENT ID to be positive, got %d", id2)
	}

	if id1 == id2 {
		t.Fatalf("expected different CLIENT IDs for different connections, both got %d", id1)
	}
}

func TestClientInfoAndList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	info, err := client.Do(ctx, "CLIENT", "INFO").Text()
	if err != nil {
		t.Fatalf("CLIENT INFO failed: %v", err)
	}
	if info == "" {
		t.Fatalf("expected CLIENT INFO to return non-empty string")
	}

	list, err := client.Do(ctx, "CLIENT", "LIST").Text()
	if err != nil {
		t.Fatalf("CLIENT LIST failed: %v", err)
	}
	if list == "" {
		t.Fatalf("expected CLIENT LIST to return non-empty string")
	}
}

func TestClientNoEvictAndNoTouch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	result, err := client.Do(ctx, "CLIENT", "NO-EVICT", "ON").Text()
	if err != nil {
		t.Fatalf("CLIENT NO-EVICT ON failed: %v", err)
	}
	if result != "OK" {
		t.Fatalf("expected CLIENT NO-EVICT ON to return OK, got %q", result)
	}

	result, err = client.Do(ctx, "CLIENT", "NO-TOUCH", "ON").Text()
	if err != nil {
		t.Fatalf("CLIENT NO-TOUCH ON failed: %v", err)
	}
	if result != "OK" {
		t.Fatalf("expected CLIENT NO-TOUCH ON to return OK, got %q", result)
	}

	err = client.Do(ctx, "CLIENT", "NO-EVICT").Err()
	if err == nil {
		t.Fatal("expected error for CLIENT NO-EVICT with no arg")
	}

	err = client.Do(ctx, "CLIENT", "NO-EVICT", "MAYBE").Err()
	if err == nil {
		t.Fatal("expected error for CLIENT NO-EVICT with invalid arg")
	}
}

func TestHello2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	// HELLO with no args
	result, err := client.Do(ctx, "HELLO").Slice()
	if err != nil {
		t.Fatalf("HELLO failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty HELLO response")
	}

	// HELLO 2 explicit
	result, err = client.Do(ctx, "HELLO", "2").Slice()
	if err != nil {
		t.Fatalf("HELLO 2 failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty HELLO 2 response")
	}
}

func TestHello3Rejected(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	err := client.Do(ctx, "HELLO", "3").Err()
	if err == nil {
		t.Fatal("expected HELLO 3 to return an error")
	}
	if !strings.Contains(err.Error(), "NOPROTO") {
		t.Fatalf("expected NOPROTO error, got %v", err)
	}
}

func TestHelloInvalidVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := startTestServer(t)
	client := newRedisClient(t, addr, 0)

	err := client.Do(ctx, "HELLO", "99").Err()
	if err == nil {
		t.Fatal("expected HELLO 99 to return an error")
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

	srv, errCh, addr := startServerWithConfig(t, cfg)
	return registerServerCleanup(t, srv, errCh, addr)
}

func startServerWithConfig(t *testing.T, cfg *server.Config) (*server.Server, <-chan error, string) {
	t.Helper()

	srv := server.NewServer(cfg)
	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start()
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if addr := srv.Addr(); addr != "" {
			return srv, errCh, addr
		}

		select {
		case err := <-errCh:
			t.Fatalf("server failed to start: %v", err)
		default:
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("server did not become ready within %s", 2*time.Second)
	return nil, nil, ""
}

func stopServer(t *testing.T, srv *server.Server, errCh <-chan error) {
	t.Helper()

	srv.Shutdown()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server shutdown returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for server shutdown")
	}
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
