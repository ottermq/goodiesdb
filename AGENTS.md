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
- `internal/core/command/`: in-progress command registry refactor.
- `internal/protocol/`: RESP abstractions and RESP2 implementation.
- `internal/persistence/aof/`: append-only file persistence and replay.
- `internal/persistence/rdb/`: snapshot persistence.

## Current branch context

Branch `feat/command_registry_pattern` is an unfinished refactor branch.

What was being built:

- A command registry that lets the server dispatch commands through command objects.
- A migration path away from the large `switch` in `internal/core/server/server.go`.
- Shared command validation and execution context.

What is already migrated:

- `GET`
- `SET`

What still uses the legacy path:

- Nearly every other command in `internal/core/server/server.go`

## Important findings

- The branch compiles, but it is not fully stabilized.
- There is a lock recursion bug in the store layer:
  - `Store.Del()` takes the store lock and then calls `delKey()`
  - `Store.Rename()` takes the store lock and then calls `delKey()`
  - `delKey()` also takes the same lock
- This causes tests to hang and timeout.

Affected files:

- `internal/core/store/store.go`
- `internal/core/store/store_utils.go`

## Known design tension

The current refactor mixes command-layer concerns and store-layer concerns:

- `Store` now holds `Protocol` so commands can emit RESP nil values.
- `GET` still contains expiration handling that partly overlaps with `Store.Get()`.

This is acceptable as a temporary bridge, but future refactor work should move protocol-aware response shaping toward the command/server layer rather than the store.

## Good first places to look

If you are resuming the command registry work, start here:

1. `internal/core/server/server.go`
2. `internal/core/command/command.go`
3. `internal/core/command/registry.go`
4. `internal/core/command/get.go`
5. `internal/core/command/set.go`
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

## Working agreements for future agents

- Do not treat Redis parity as the main goal; client compatibility is the main goal.
- Avoid destructive git operations.
- Prefer small, reviewable refactor slices.
- Keep docs updated when the branch intent changes.
- If touching command dispatch, preserve backward behavior unless intentionally changing semantics.
