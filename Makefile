MAELSTROM_BIN=./maelstrom/maelstrom

maelstrom:
	wget https://github.com/jepsen-io/maelstrom/releases/download/v0.2.4/maelstrom.tar.bz2
	tar -xvjf ./maelstrom.tar.bz2
	rm ./maelstrom.tar.bz2

01-echo:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w echo --bin ./$@/build --node-count 1 --time-limit 10
.PHONY: 01-echo

02-unique-id-generation:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w unique-ids --bin ./$@/build --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
.PHONY: 02-unique-id-generation

3a-single-node-broadcast:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w broadcast --bin ./$@/build --node-count 1 --time-limit 20 --rate 10
.PHONY: 3a-single-node-broadcast

3b-multi-node-broadcast:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w broadcast --bin ./$@/build --node-count 5 --time-limit 20 --rate 10
.PHONY: 3b-multi-node-broadcast

3c-fault-tolerant-broadcast:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w broadcast --bin ./$@/build --node-count 5 --time-limit 20 --rate 10 --nemesis partition
.PHONY: 3c-fault-tolerant-broadcast

4-grow-only-counter:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w g-counter --bin ./$@/build --node-count 3 --rate 100 --time-limit 20 --nemesis partition
.PHONY: 4-grow-only-counter

5a-single-node-kafka-style-log:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w kafka --bin ./$@/build --node-count 1 --concurrency 2n --time-limit 20 --rate 1000
.PHONY: 5a-single-node-kafka-style-log

5b-multi-node-kafka-style-log:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w kafka --bin ./$@/build --node-count 2 --concurrency 2n --time-limit 20 --rate 1000
.PHONY: 5b-multi-node-kafka-style-log

6a-single-node-totally-available-transactions:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w txn-rw-register --bin ./$@/build --node-count 1 --time-limit 20 --rate 1000 --concurrency 2n --consistency-models read-uncommitted --availability total
.PHONY: 6a-single-node-totally-available-transactions

6b-totally-available-read-uncommitted-transactions:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w txn-rw-register --bin ./$@/build --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted --availability total --nemesis partition
.PHONY: 6b-totally-available-read-uncommitted-transactions

6c-totally-available-read-committed-transactions:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w txn-rw-register --bin ./$@/build --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-committed --availability total --nemesis partition
.PHONY: 6c-totally-available-read-committed-transactions
