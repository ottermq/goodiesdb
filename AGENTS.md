# GoodiesDB Agent Orientation

This file is for Codex agents and future contributors who need fast context before making changes.

## Project intent

GoodiesDB started as a Redis clone, but the current goal is narrower and more practical:

- Be compatible with popular Redis clients where it is useful.
- Be mature enough to replace Redis in the author's pet projects.
- Stay educational: the codebase is a vehicle for learning Redis design tradeoffs.
- It is not trying to become a production Redis competitor.

When making decisions, prefer:

- Redis client compatibility over perfect internal elegance.
- Incremental improvements over large rewrites.
- Clear behavior and testability over feature count.

## Current architecture

- `cmd/goodiesdb-server/main.go`: process startup, config loading, graceful shutdown.
- `internal/core/server/`: TCP server, RESP request handling, per-connection state, command dispatch.
- `internal/core/store/`: in-memory multi-database data store and data-type operations.
- `internal/core/command/`: command implementations and shared validation/RESP helpers.
- `internal/protocol/`: RESP abstractions and RESP2 implementation.
- `internal/persistence/aof/`: append-only file persistence and replay.
- `internal/persistence/rdb/`: snapshot persistence.

## Current branch context

Branch `feat/command_registry_pattern` is an unfinished refactor branch.

What was being built:

- A command registry that lets the server dispatch commands through command objects.
- A migration path away from the large `switch` in `internal/core/server/server.go`.
- Shared command validation and execution context.

Current status:

- The command registry migration is complete for the currently implemented command set.
- The legacy command switch has been removed.
- Command validation now lives in `Validate()` with shared helpers in the command package.

## Important findings

- The original branch-level deadlock in the store delete path has been fixed.
- The store no longer knows about protocol encoding.
- Nil reply shaping now belongs to the command/server layer, not the store.

## Known design tension

- `GET` still contains expiration handling that partly overlaps with `Store.Get()`.
- Connection-aware behavior still depends directly on `net.Conn` and callbacks in command context.

Those are the main remaining cleanup areas after the registry migration itself.

## Good first places to look

If you are resuming the command registry work, start here:

1. `internal/core/server/server.go`
2. `internal/core/command/command.go`
3. `internal/core/command/registry.go`
4. `internal/core/command/validate_helpers.go`
5. `internal/core/command/resp_helpers.go`
6. `docs/PROJECT_STATUS.md`
7. `docs/refactoring/COMMAND_REGISTRY_REFACTOR.md`

## Practical workflow

- Read `README.md` for project framing.
- Read `.github/copilot-instructions.md` for repo-specific guidance.
- Prefer `rg` for navigation.
- Run tests with a writable Go cache when sandboxed:

```bash
GOCACHE=/tmp/gocache go test ./...
```

## Testing strategy

GoodiesDB was historically validated manually by connecting with a Redis client and trying commands interactively. That revealed the most important testing constraint in this project:

- the behavior that matters most is the behavior seen by Redis clients

Because of that, the preferred testing strategy going forward is layered:

- unit tests for store-level behavior and low-level primitives
- integration tests that boot GoodiesDB and exercise it through Redis client libraries

The integration layer should become the main compatibility safety net because it verifies:

- request parsing
- reply encoding
- command semantics
- nil and error behavior
- database selection and connection behavior

## Working agreements for future agents

- Do not treat Redis parity as the main goal; client compatibility is the main goal.
- Avoid destructive git operations.
- Prefer small, reviewable refactor slices.
- Keep docs updated when the branch intent changes.
- If touching command dispatch, preserve backward behavior unless intentionally changing semantics.
