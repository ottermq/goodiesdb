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

Status now:

- the currently implemented command set has been migrated into the registry
- the legacy switch-based execution path has been removed
- unknown commands now fail directly after registry lookup

The incremental migration strategy worked and is now complete for this phase.

## Current problems to resolve

### 1. Historical baseline regression

The branch originally had a store deadlock in the delete path. That regression has been fixed, and the test suite is green again.

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

## What remains after migration

The main remaining work is cleanup and hardening, not command extraction.

Recommended follow-ups:

- simplify the `Command` interface if `MinArgs()` and `MaxArgs()` remain unused
- decide whether `Store.Protocol` should remain in the store layer
- tighten session-aware behavior where raw `net.Conn` still feels awkward
- add more compatibility tests around errors, edge cases, and unsupported commands
- remove or reshape any transitional APIs that only existed to ease migration

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
