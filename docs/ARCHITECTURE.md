# Architecture Overview

## Runtime flow

At a high level, GoodiesDB works like this:

1. `cmd/goodiesdb-server/main.go` loads config and starts the server
2. `internal/core/server/` accepts TCP connections
3. RESP requests are parsed through `internal/protocol/resp2/`
4. The server resolves the command and active database
5. Command logic operates on `internal/core/store/`
6. Persistence is handled through AOF and/or RDB when enabled

## Main modules

### Server

`internal/core/server/` owns:

- connection lifecycle
- authentication tracking
- selected database per connection
- command dispatch
- startup and recovery flow

### Store

`internal/core/store/` owns:

- in-memory data per logical database
- type-aware values
- key expiration metadata
- primitive operations used by commands

### Command layer

`internal/core/command/` owns command-specific execution logic, validation, and a small set of shared command helpers.

This layer is now the primary command-dispatch mechanism.

### Protocol

`internal/protocol/` defines RESP values and the RESP2 implementation used by the server.

### Persistence

`internal/persistence/aof/` appends write commands to disk as RESP arrays and can replay them losslessly.

`internal/persistence/rdb/` saves and restores snapshots of store state.

## Important implementation notes

- The store currently supports 16 logical databases.
- Database selection is per connection.
- Persistence is optional and configured through server config.
- AOF recovery only supports the current RESP-based format; legacy line-based AOF files are ignored.
- Command dispatch is now registry-based rather than switch-based.
- Nil reply encoding is handled at the command/server layer, not inside the store.

## Testing philosophy

The preferred testing model for GoodiesDB is layered:

- unit tests validate internal store and persistence behavior
- integration tests validate end-to-end compatibility through Redis client libraries

This is intentional. GoodiesDB's primary promise is client compatibility for practical pet-project use, so tests should exercise the system the way real consumers use it.

That means a feature is best validated by:

1. starting the server
2. connecting with a Redis client library
3. asserting behavior through the network protocol

This should gradually replace the earlier manual-only workflow of opening a Redis client and trying commands by hand.

## Architectural debt worth watching

- connection-aware behavior still depends directly on `net.Conn`
- command context currently carries server callbacks for behaviors like auth and DB selection
- connection state is tracked directly on `net.Conn`, which may become awkward as command behavior grows
