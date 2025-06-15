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