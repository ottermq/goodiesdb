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

Since then, the branch has been continued and the migration has been completed for the currently implemented commands.

## Current command execution model

The command registry migration is now effectively complete for the currently implemented command set.

Today the server uses this flow:

- parse the RESP array
- resolve `cmdName`
- load the command from `commandRegistry`
- execute it through `invokeCommand()`
- return `ERR unknown command '<name>'` if no command is registered

The old command switch has been retired.

## Compatibility progress

The current branch now includes a first useful slice of Redis hash support aimed at real client flows:

- `HSET`
- `HGET`
- `HGETALL`
- `HDEL`
- `HEXISTS`
- `HLEN`
- `HMGET`
- `HKEYS`
- `HVALS`

This moves hashes from "planned" toward "usable for common application state", especially for object and session-style storage.

## Known issues

### Historical note: deadlock in store deletion path

The branch originally contained a lock recursion bug in the delete path. That issue has been fixed.

### Transitional layering smell

That bridge has now been removed. Nil reply shaping is handled by the command/server layer instead of the store.

### Responsibility overlap

`GET` command still performs explicit expiration handling even though `Store.Get()` already suppresses expired values.

## What should happen next

Recommended order:

1. Strengthen integration coverage for edge cases and unsupported command behavior.
2. Continue improving Redis-client compatibility command by command.
3. Revisit whether session-aware commands need a richer connection abstraction.
4. Clarify expiration responsibility between command and store layers.
5. Keep trimming transitional abstractions where they no longer pay for themselves.

## Testing direction

Historically, features were validated manually by connecting with a Redis client and checking behavior interactively.

That should now become an explicit automated testing strategy:

- keep unit tests for store and persistence behavior
- add integration tests that start GoodiesDB and use Redis client libraries to exercise features end to end

This is the right fit for the project because the main success criterion is not strict internal design purity. It is whether Redis clients can use GoodiesDB without surprises.

Priority test targets:

- `SET` and `GET`
- expiration behavior
- list operations
- `SELECT`
- nil responses
- error replies
- command argument validation

## Resume checklist

When restarting implementation work:

1. Read `AGENTS.md`
2. Read `docs/refactoring/COMMAND_REGISTRY_REFACTOR.md`
3. Re-run `GOCACHE=/tmp/gocache go test ./...`
4. Review the remaining architectural cleanup items
5. Extend compatibility tests for edge cases before large behavior changes
