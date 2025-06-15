package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const counterName = "counter"

func main() {
	node := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(node)

	s := server{
		node: node,
		kv:   kv,
	}

	node.Handle("add", s.handleAdd)
	node.Handle("read", s.handleRead)

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	node *maelstrom.Node
	kv   *maelstrom.KV
}

func (s *server) handleAdd(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	delta := int(body["delta"].(float64))

	for {
		counterVal, err := s.kv.ReadInt(context.TODO(), counterName)
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				counterVal = 0
			} else {
				return err
			}
		}

		err = s.kv.CompareAndSwap(context.TODO(), counterName, counterVal, counterVal+delta, true)
		if err == nil {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	response := map[string]any{
		"type": "add_ok",
	}

	return s.node.Reply(msg, response)
}

func (s *server) handleRead(msg maelstrom.Message) error {
	var counterVal int
	var err error

	for {
		counterVal, err = s.readCounterWithSynchronization()
		if err != nil {
			time.Sleep(100 * time.Second)
			continue
		}

		break
	}

	response := map[string]any{
		"type":  "read_ok",
		"value": counterVal,
	}

	return s.node.Reply(msg, response)
}

// reads the counter and confirms its value is stable
func (s *server) readCounterWithSynchronization() (int, error) {
	counterVal, err := s.kv.ReadInt(context.TODO(), counterName)
	if err != nil {
		if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
			return 0, nil
		} else {
			return 0, err
		}
	}

	// Use a no-op CAS to ensure consistency
	err = s.kv.CompareAndSwap(context.TODO(), counterName, counterVal, counterVal, false)
	if err != nil {
		return 0, err
	}

	return counterVal, nil
}
