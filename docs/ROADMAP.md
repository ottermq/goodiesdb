# Roadmap

## Product direction

GoodiesDB should become a practical Redis-compatible datastore for personal projects, while remaining intentionally small and understandable.

The roadmap is guided by three questions:

- Can common Redis clients talk to it without surprises?
- Is the behavior reliable enough for small real projects?
- Does the implementation stay teachable?

## Near term

### 1. Stabilize the current branch

- fix the store deadlock introduced during the command registry refactor
- get `go test ./...` passing again
- document the refactor boundaries and intent

### 2. Finish the command registry migration foundation

- strengthen the `Command` abstraction
- decide where argument validation belongs
- reduce protocol coupling inside `Store`
- add tests for registry-based dispatch

### 3. Migrate easy commands first

Best early candidates:

- `DEL`
- `EXISTS`
- `EXPIRE`
- `INCR`
- `DECR`
- `TTL`
- `GETRANGE`
- `STRLEN`
- `TYPE`
- `KEYS`
- `PING`
- `ECHO`

These commands are relatively self-contained and should help prove the pattern.

## Mid term

### 4. Improve client-facing compatibility

- match Redis argument validation more consistently
- align reply types and nil behavior
- verify behavior with popular Redis clients
- identify unsupported but commonly expected commands

### 5. Untangle connection-level behavior

- revisit auth handling
- create a clearer per-connection/session abstraction
- move `AUTH`, `SELECT`, and `QUIT` into a more structured execution path

### 6. Strengthen persistence confidence

- add better coverage around AOF replay semantics
- verify persistence with more command types
- define expected precedence and recovery behavior more explicitly

## Longer term

### 7. Expand useful data structures carefully

Add features only when they improve practical usability or reveal important implementation lessons, such as:

- hashes
- sets
- sorted sets
- pub/sub
- transactions

### 8. Deepen operational realism

Areas worth exploring later:

- expiration cleanup strategy
- protocol/version coverage
- replication concepts
- snapshot and AOF tradeoffs
- memory behavior and observability

## Non-goals for now

- competing with Redis on performance
- full Redis feature parity
- production-grade distributed systems concerns

## Success criteria

GoodiesDB is on track when:

- common Redis client flows work cleanly
- tests are reliable and reasonably fast
- command behavior is easy to extend without touching a giant switch
- the code still feels understandable to a single developer returning after a long break
