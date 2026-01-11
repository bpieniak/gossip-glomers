# Gossip Glomers

My attempt at [Gossip Glomers](https://fly.io/dist-sys/), a series of distributed systems challenges.

## Solutions

### Challenge #2: Unique ID Generation

[Solution](02-unique-id-generation/main.go)

Just used `uuid.New()` from `github.com/google/uuid` package to generate unique ids.

### Challenge #3: Broadcast

#### 3a-single-node-broadcast

[Solution](3a-single-node-broadcast/main.go)

Simple case - stores incoming messages in a local list. There’s no communication with other nodes.

#### 3a-single-node-broadcast

[Solution](3b-multi-node-broadcast/main.go)

When a message is received, it's stored locally and forwarded to neighboring nodes based on the provided topology.

#### 3c-fault-tolerant-broadcast

[Solution](3c-fault-tolerant-broadcast/main.go)

Expansion of 3a. Messages are stored in a map for deduplication and messages are forwarded to neighbors with retry logic that ensures that messages eventually reach all nodes.

#### 3d-efficient-broadcast-part-1

...

###  Challenge #4: Grow-Only Counter

[Solution](4-grow-only-counter/main.go)

In this implementation in `add` use CAS to atomically update counter in SeqKV, retry on failure. `read` is done by reading counter from SeqKV and performing a no-op CAS to ensure the value is stable - if CAS fails (due to a concurrent write), retry the entire read.

### Challenge #5: Kafka-Style Log

#### Challenge #5a: Single-Node Kafka-Style Log

[Solution](5a-single-node-kafka-style-log/main.go)

Implements a single-node, per-key append-only log with monotonic offsets. State lives in-memory per key, protected by a mutex. Each log tracks its message slice and latest committed offset to satisfy Maelstrom’s ordering and loss checks.

#### Challenge #5b: Multi-Node Kafka-Style Log

[Solution](5b-multi-node-kafka-style-log/main.go)

Distributes the log across nodes using Maelstrom’s linearizable KV: per key we CAS a `next` counter to allocate offsets, write messages under keyed offset entries, store commits separately, and serve polls by reading stored offsets.

### Totally-Available Transaction

#### Challenge #6a: Single-Node, Totally-Available Transactions

[Solution](6a-single-node-totally-available-transactions/main.go)

Handles `txn` requests on a single node with an in-memory key/value map guarded by a mutex. The handler applies each read/write in order and echoes back the transaction with read results, which is sufficient for the single-node, totally-available case.

#### Challenge #6b: Totally-Available Read Uncommitted Transactions

[Solution](6b-totally-available-read-uncommitted-transactions/main.go)

Replicates write operations to all nodes using best-effort async sends while serving reads from local state for a read-uncommitted model. Each txn applies reads/writes locally and replies immediately to preserve total availability.
