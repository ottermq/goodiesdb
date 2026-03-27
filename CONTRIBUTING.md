# Contributing to GoodiesDB

## Purpose

GoodiesDB is a learning-driven Redis-compatible datastore written in Go. Contributions should help the project become easier to understand, safer to evolve, and more compatible with common Redis client behavior.

## Development goals

Prioritize work in this order:

1. Correctness and predictable behavior
2. Compatibility with common Redis clients
3. Code clarity and maintainability
4. New features

This repository is intentionally educational. Please prefer straightforward implementations that are easy to reason about.

## Local workflow

Build:

```bash
make build
```

Run:

```bash
make run
```

Test:

```bash
go test ./...
```

If you are running inside a sandboxed environment, use a writable build cache:

```bash
GOCACHE=/tmp/gocache go test ./...
```

## Areas of the codebase

- `internal/core/server/`: networking, command dispatch, connection state
- `internal/core/store/`: in-memory data model and command primitives
- `internal/core/command/`: in-progress command registry refactor
- `internal/persistence/`: AOF and RDB persistence
- `internal/protocol/`: RESP protocol representation and encoding

## Contribution guidelines

- Keep changes small and well scoped.
- Add or update tests when behavior changes.
- Prefer fixing regressions before adding features.
- Document architectural intent when introducing refactors.
- Preserve Redis-like behavior when practical, but do not add complexity purely for mimicry.

## Current refactor note

The branch `feat/command_registry_pattern` is in the middle of extracting command handling from the server switch statement into command objects.

If you are contributing on or after this branch, read these first:

- `AGENTS.md`
- `docs/PROJECT_STATUS.md`
- `docs/ROADMAP.md`
- `docs/refactoring/COMMAND_REGISTRY_REFACTOR.md`

## Testing expectations

At minimum, verify the areas you changed.

Examples:

- Store changes: `go test ./internal/core/store/...`
- Persistence changes: `go test ./internal/persistence/...`
- Cross-cutting behavior: `go test ./...`

When possible, validate behavior from a Redis client as well, especially for argument validation and reply shape.

## Style notes

- Follow existing Go formatting conventions.
- Prefer explicit, readable code over clever abstractions.
- Avoid introducing new layers unless they reduce complexity across multiple commands.
- Keep command semantics close to Redis where it improves compatibility.

## Documentation

If your change affects architecture, project direction, or refactor strategy, update the relevant docs in `docs/`.
