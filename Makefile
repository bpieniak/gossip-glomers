MAELSTROM_BIN=./maelstrom/maelstrom

01-echo:
	go build -o ./$@/build ./$@
	${MAELSTROM_BIN} test -w echo --bin ./$@/build --node-count 1 --time-limit 10
.PHONY: 01-echo
