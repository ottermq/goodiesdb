# Project Status

## Snapshot

Last updated: April 2026.

## What GoodiesDB is trying to be

A Redis-compatible datastore good enough for pet projects — not a production Redis
replacement, but a practical and educational implementation that real Redis clients
can talk to without surprises.

## Architecture

- **Command registry** — all commands registered as dedicated types, dispatched
  through `invokeCommand`. No monolithic switch.
- **Connection abstraction** — each client connection is tracked as a `Conn` struct
  holding DB index, auth status, and pub/sub mode.
- **Store** — in-memory, 16 databases, lazy expiration, decoupled from protocol
  encoding.
- **Persistence** — AOF (RESP-encoded, lossless replay) and RDB snapshots. Both
  optional and independently configurable.
- **Protocol** — RESP2. RESP3 types defined but not yet used.
- **Pub/sub broker** — global (across all DBs), one delivery channel per connection,
  exact and glob-pattern fanout, drop-on-full with logging.

## Implemented commands

### Strings
`GET` `SET` `SETNX` `INCR` `DECR` `STRLEN` `GETRANGE`

### Lists
`LPUSH` `RPUSH` `LPOP` `RPOP` `LRANGE` `LTRIM`

### Hashes
`HSET` `HGET` `HGETALL` `HDEL` `HEXISTS` `HLEN` `HMGET` `HKEYS` `HVALS`

### Key management
`DEL` `EXISTS` `EXPIRE` `TTL` `TYPE` `KEYS` `RENAME` `SCAN`

### Pub/Sub
`SUBSCRIBE` `UNSUBSCRIBE` `PUBLISH` `PSUBSCRIBE` `PUNSUBSCRIBE`

### Server
`AUTH` `SELECT` `INFO` `PING` `ECHO` `QUIT` `FLUSHDB` `FLUSHALL`

## Pub/Sub notes

- Channels are global across all databases.
- Subscriber mode is enforced — only sub/unsub/ping/quit allowed after `SUBSCRIBE`.
- `PING` in subscriber mode returns the push-format pong array, not `+PONG`.
- Pattern matching uses glob semantics (`path.Match`), same as Redis.
- Disconnect cleanup is automatic — the broker delivery channel is closed, which
  terminates the write goroutine.
- Known gap: `UNSUBSCRIBE` with zero args sends a single nil-channel confirmation
  instead of one per subscribed channel. Redis sends one per channel.

## Testing

Integration tests in `integration/` start GoodiesDB and exercise it through the
real `go-redis/v9` client. Unit tests cover store and persistence internals.
Broker behaviour is tested in `internal/core/server/pubsub_test.go`.

Run all tests:
```
go test ./...
```

## Known issues / next steps

1. `UNSUBSCRIBE` zero-args gap — broker needs a method to list channels for a conn.
2. Client-facing compatibility — argument validation and nil/error reply alignment
   still has gaps; work command by command against real client expectations.
3. Connection-level commands — `AUTH`, `SELECT`, `QUIT` should route through the
   `Conn` abstraction more cleanly.
4. Expiration responsibility — `GET` still does explicit expiration handling even
   though `Store.Get()` already suppresses expired values.
5. Sets and sorted sets — not yet implemented.

## Resume checklist

1. Run `go test ./...` to verify green baseline.
2. Check `docs/ROADMAP.md` for current priorities.
3. Pick up from the known issues list above.
