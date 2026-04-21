# Goodies DB - A Redis implementation in Go

[![Go](https://github.com/andrelcunha/goodiesdb/actions/workflows/go.yml/badge.svg)](https://github.com/andrelcunha/goodiesdb/actions/workflows/go.yml)
[![Docker Image CI](https://github.com/andrelcunha/goodiesdb/actions/workflows/docker-image.yml/badge.svg)](https://github.com/andrelcunha/goodiesdb/actions/workflows/docker-image.yml)

GoodiesDb started as a Redis implementation written in Go, serving as an educational project to learn and understand the inner workings of Redis, a popular in-memory data structure store. The current state of the project implements a subset of Redis commands across strings, lists, hashes, key management, pub/sub, and server operations — see [Features](#features) for the full list.

**Disclaimer:** This is not a production-ready Redis clone and it is not intended for use in production environments (yet).

---

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Introduction

GoodiesDB aims to mimic the basic functionalities of Redis to provide a learning platform for developers interested in understanding distributed systems, data structures, and high-performance computing.

## Features

- In-memory key-value store with 16 logical databases
- Data persistence via AOF (RESP-encoded) and RDB snapshots
- Registry-based command dispatch — no monolithic switch
- Pub/Sub messaging with pattern matching and subscriber mode enforcement

### Supported commands

**Strings:** `SET` `GET` `SETNX` `INCR` `DECR` `STRLEN` `GETRANGE`

**Lists:** `LPUSH` `RPUSH` `LPOP` `RPOP` `LRANGE` `LTRIM`

**Hashes:** `HSET` `HGET` `HGETALL` `HDEL` `HEXISTS` `HLEN` `HMGET` `HKEYS` `HVALS`

**Key management:** `DEL` `EXISTS` `EXPIRE` `TTL` `TYPE` `KEYS` `RENAME` `SCAN`

**Pub/Sub:** `SUBSCRIBE` `UNSUBSCRIBE` `PSUBSCRIBE` `PUNSUBSCRIBE` `PUBLISH`

**Server:** `AUTH` `SELECT` `INFO` `PING` `ECHO` `QUIT` `FLUSHDB` `FLUSHALL`

## Installation

To get started with GoodiesDB, follow these steps:

1. **Clone the repository**:

    ```bash
    git clone https://github.com/andrelcunha/goodiesdb.git
    cd GoodiesDB
    ```

2. **Install dependencies**:

    ```bash
    go mod tidy
    ```

3. **Build the project**:

    ```bash
    make build
    ```

## Usage

Run the GoodiesDb server:

```bash
make run
```

Control log verbosity with `LOG_LEVEL`:

```bash
LOG_LEVEL=debug make run
```

Supported values are `error`, `info`, and `debug`. The default is `info`.

You can then connect with any Redis-compatible client on port 6379, or run the integration test suite:

```bash
go test ./...
```

## Documentation

For project context and contributor guidance, start with:

- `AGENTS.md`
- `CONTRIBUTING.md`
- `docs/ARCHITECTURE.md`
- `docs/PROJECT_STATUS.md`
- `docs/ROADMAP.md`
- `docs/TESTING.md`
- `docs/refactoring/COMMAND_REGISTRY_REFACTOR.md`

## Persistence note

GoodiesDB now writes AOF files in RESP format rather than the older line-based space-split format.

Older `appendonly.aof` files from the legacy format are not replayed by current versions. If such a file is present, GoodiesDB starts with an empty store instead of attempting a best-effort import.

## License

This project is licensed under the MIT License.

## Acknowledgements

- [Redis](https://redis.io/documentation) for the inspiration and original implementation.
- [Golang](https://golang.org/) for the programming language.
