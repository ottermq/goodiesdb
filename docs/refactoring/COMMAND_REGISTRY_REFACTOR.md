# Command Registry Refactor Guide

## Why this refactor exists

Before this branch, command execution lived almost entirely inside a large `switch` in `internal/core/server/server.go`.

That approach works for a small toy project, but it starts to hurt once the goals become:

- grow command coverage
- improve Redis compatibility
- validate command behavior consistently
- keep the codebase understandable as it matures

The command registry refactor is intended to make command execution modular and incremental.

## Current design

The refactor introduces:

- `Command` interface in `internal/core/command/command.go`
- `Context` object passed to commands
- `Registry` for command lookup
- `Server.invokeCommand()` as a bridge from the network layer to command objects

Current migrated commands:

- `GET`
- `SET`

Current migration strategy:

- register a command in the registry
- let the server try the registry first
- keep the old switch as fallback
- migrate commands gradually

This is a sensible strategy and should be preserved.

## Current problems to resolve

### 1. Baseline regression

Fix the store deadlock before continuing the refactor.

Without a stable baseline, it is too easy to confuse refactor mistakes with pre-existing behavior.

### 2. Command interface shape is not settled

Current interface:

- `Name()`
- `Execute()`
- `Validate()`
- `RequiresAuth()`
- `MinArgs()`
- `MaxArgs()`

Observations:

- `MinArgs()` and `MaxArgs()` are currently unused.
- `SET` supports variable options, so `MaxArgs()` is already a poor fit as written.
- `Validate()` is doing the real work today.

Recommendation:

- either remove `MinArgs()` and `MaxArgs()`
- or make the server use them consistently for generic validation

Do not keep redundant abstractions unless they reduce code meaningfully.

### 3. Protocol concerns leaked into the store

`Store` currently has a `Protocol` field so commands can return protocol-specific nil values.

Recommendation:

- keep this only if necessary as a short-term bridge
- prefer moving response-shaping helpers into command context or server utilities

The store should ideally stay focused on data and behavior, not wire format concerns.

### 4. Expiration ownership is blurry

`Store.Get()` already returns missing for expired keys, but `GET` still performs explicit expiration handling and lazy deletion.

Recommendation:

- decide whether expiration cleanup belongs in the store or command layer
- make that rule explicit
- apply it consistently

## Recommended migration order

### Phase 1: restore safety

- fix deadlock in delete path
- re-run tests
- add a small regression test if needed

### Phase 2: prove the pattern

Migrate a small group of simple commands:

- `DEL`
- `EXISTS`
- `EXPIRE`
- `INCR`
- `DECR`
- `TTL`

Why these first:

- simple argument shapes
- minimal connection-state coupling
- easy to compare against existing switch behavior

### Phase 3: expand utility commands

Next candidates:

- `GETRANGE`
- `STRLEN`
- `TYPE`
- `KEYS`
- `PING`
- `ECHO`

### Phase 4: handle session-aware commands

Later candidates:

- `AUTH`
- `SELECT`
- `QUIT`

These likely deserve a richer connection/session abstraction than raw `net.Conn`.

### Phase 5: move complex list commands

After the pattern is stable:

- `LPUSH`
- `RPUSH`
- `LPOP`
- `RPOP`
- `LRANGE`
- `LTRIM`

### Phase 6: remove legacy switch paths

Only after enough migration and test coverage:

- shrink the switch
- keep only bootstrap or truly special commands if needed
- eventually retire the switch-based command flow

## What good looks like

This refactor is successful when:

- adding a command does not require editing a giant switch
- validation and reply shaping are consistent
- connection-aware behavior is explicit
- store logic remains mostly protocol-agnostic
- tests cover both command behavior and dispatch

## Suggested guardrails while refactoring

- migrate commands in small batches
- run tests after each batch
- preserve existing behavior unless intentionally improving compatibility
- keep docs current when the design changes
- avoid mixing command migration with unrelated feature work
