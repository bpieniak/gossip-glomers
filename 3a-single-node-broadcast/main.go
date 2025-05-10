package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	s := server{
		node: n,
	}

	n.Handle("broadcast", s.handleBroadcast)
	n.Handle("read", s.handleRead)
	n.Handle("topology", s.handleTopology)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	node *maelstrom.Node

	valuesMu sync.Mutex
	values   []float64
}

func (s *server) handleBroadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.valuesMu.Lock()
	s.values = append(s.values, body["message"].(float64))
	s.valuesMu.Unlock()

	response := map[string]any{
		"type": "broadcast_ok",
	}

	return s.node.Reply(msg, response)
}
func (s *server) handleRead(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.valuesMu.Lock()
	response := map[string]any{
		"type":     "read_ok",
		"messages": s.values,
	}
	s.valuesMu.Unlock()

	return s.node.Reply(msg, response)
}

func (s *server) handleTopology(msg maelstrom.Message) error {
	response := map[string]any{
		"type": "topology_ok",
	}

	return s.node.Reply(msg, response)
}
