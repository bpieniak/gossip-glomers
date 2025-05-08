MAELSTROM_BIN=./maelstrom/maelstrom

01-echo:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w echo --bin ./$@/build --node-count 1 --time-limit 10
.PHONY: 01-echo

02-unique-id-generation:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w unique-ids --bin ./$@/build --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
.PHONY: 02-unique-id-generation