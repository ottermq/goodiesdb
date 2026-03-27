# Project Status

## Snapshot

This document captures the state of the repository as of the `feat/command_registry_pattern` branch analysis on March 27, 2026.

## What GoodiesDB is trying to be

GoodiesDB is no longer just a Redis clone exercise. The practical target is:

- good enough Redis compatibility for pet projects
- enough maturity to serve as a Redis substitute in those projects
- an educational codebase that exposes the kinds of problems Redis developers had to solve

## What the branch was doing

The current branch appears to be an incremental refactor toward a command registry pattern.

The likely objectives were:

- remove the monolithic command switch from the server
- encapsulate command behavior in dedicated types
- centralize command validation
- make future command additions and compatibility work easier

## Evidence from commit history

Recent branch commits, in order:

1. add `Command` interface and `Context`
2. add command registry
3. implement `GET` command
4. add `Store.Delete()`
5. add `Store.SetProtocol()`
6. add `Value.ToString()`
7. integrate registry into server
8. implement `SET` command

This strongly suggests the refactor stopped shortly after a vertical slice for `GET` and `SET`.

## Current command execution model

Today the server uses a hybrid approach:

- first try `commandRegistry.Get(cmdName)`
- if a command is registered, execute it through `invokeCommand()`
- otherwise fall back to the legacy `switch` in `internal/core/server/server.go`

Current registry coverage:

- `GET`
- `SET`

Legacy server switch still handles:

- `AUTH`
- `DEL`
- `EXISTS`
- `SETNX`
- `EXPIRE`
- `INCR`
- `DECR`
- `TTL`
- `SELECT`
- list commands
- `RENAME`
- `TYPE`
- `KEYS`
- `INFO`
- `PING`
- `ECHO`
- `QUIT`
- `FLUSHDB`
- `FLUSHALL`
- `SCAN`
- `GETRANGE`
- `STRLEN`

## Known issues

### Deadlock in store deletion path

There is a lock recursion bug introduced on this branch.

Problem:

- `Store.Del()` acquires `s.mu.Lock()` and then calls `delKey()`
- `Store.Rename()` acquires `s.mu.Lock()` and then calls `delKey()`
- `delKey()` also acquires `s.mu.Lock()`

Result:

- store tests hang
- AOF tests hang
- full test runs timeout

### Transitional layering smell

`Store` currently carries a `Protocol` field so commands can emit RESP nil replies. This is workable as a bridge, but it mixes storage and protocol concerns.

### Responsibility overlap

`GET` command still performs explicit expiration handling even though `Store.Get()` already suppresses expired values.

## What should happen next

Recommended order:

1. Fix the deadlock and restore a passing baseline.
2. Add tests for the command registry path.
3. Clarify command-layer responsibilities.
4. Migrate simple stateless commands out of the server switch.
5. Delay connection-state commands until there is a cleaner session abstraction.

## Resume checklist

When restarting implementation work:

1. Read `AGENTS.md`
2. Read `docs/refactoring/COMMAND_REGISTRY_REFACTOR.md`
3. Fix the lock recursion first
4. Re-run `GOCACHE=/tmp/gocache go test ./...`
5. Continue migration in small command batches
