# Testing Strategy

## Goal

The goal of testing in GoodiesDB is confidence in real client-visible behavior.

GoodiesDB is not mainly trying to imitate Redis internally. It is trying to behave well enough that common Redis clients can use it successfully in pet projects.

Because of that, tests should focus on externally observable behavior.

## Testing layers

### 1. Unit tests

Use unit tests for:

- store primitives
- persistence helpers
- data structure edge cases
- regressions that are easiest to isolate without a running server

Unit tests should stay fast and focused.

### 2. Integration tests through Redis client libraries

Use integration tests for:

- end-to-end command behavior
- request parsing and response encoding
- nil handling
- error replies
- connection-scoped behavior
- selected database behavior
- compatibility with common Redis client expectations

These tests should start a GoodiesDB server and communicate with it using a Redis client library rather than by calling internal methods directly.

This layer should become the main regression safety net for user-visible behavior.

## Why client-library tests matter here

GoodiesDB was historically tested manually by opening a Redis client and trying commands.

That revealed the most important truth about the project:

- a feature is only truly useful if it works correctly through a Redis client

Automating that workflow gives us:

- repeatability
- better confidence during refactors
- protection against protocol and reply-shape regressions
- a clearer definition of compatibility

## Suggested initial coverage

Start with a small but high-value suite:

- `SET` and `GET`
- missing keys and nil responses
- `EXPIRE` and TTL-related behavior
- `INCR` and `DECR`
- list push/pop operations
- `SELECT`
- wrong argument count handling
- wrong type errors

## Practical guidance

- Prefer feature-oriented tests over file-oriented tests.
- One integration test can validate the whole request-to-store-to-response path.
- Keep unit tests when they make failures easier to localize.
- Do not remove low-level tests just because a higher-level test exists.

## Execution notes

In sandboxed environments, Go may need a writable build cache:

```bash
GOCACHE=/tmp/gocache go test ./...
```

## Decision rule

When adding a new behavior or fixing a bug, ask:

- should this be protected by a unit test?
- should this also be protected by an end-to-end Redis client test?

In most user-visible command changes, the answer to the second question should be yes.
