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

