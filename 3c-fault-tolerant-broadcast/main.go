package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	s := server{
		node:   n,
		values: map[float64]interface{}{},
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

	topology map[string][]string

	valuesMu sync.Mutex
	values   map[float64]interface{}
}

func (s *server) handleBroadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	message := body["message"].(float64)

	s.valuesMu.Lock()
	if _, exists := s.values[message]; exists {
		return s.node.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})
	}
	s.values[message] = struct{}{}
	s.valuesMu.Unlock()

	// Broadcast to neighbors
	for _, nodeId := range s.topology[s.node.ID()] {
		if nodeId == msg.Src {
			continue
		}

		s.broadcast(nodeId, message)
	}

	return s.node.Reply(msg, map[string]any{
		"type": "broadcast_ok",
	})
}

func (s *server) broadcast(dest string, message float64) {
	body := map[string]any{
		"type":    "broadcast",
		"message": message,
	}

	var i time.Duration
	for i = 1; ; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()

		_, err := s.node.SyncRPC(ctx, dest, body)
		if err == nil {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (s *server) handleRead(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.valuesMu.Lock()
	valuesSlice := make([]float64, 0, len(s.values))
	for value := range s.values {
		valuesSlice = append(valuesSlice, value)
	}
	s.valuesMu.Unlock()

	response := map[string]any{
		"type":     "read_ok",
		"messages": valuesSlice,
	}

	return s.node.Reply(msg, response)
}

func (s *server) handleTopology(msg maelstrom.Message) error {
	var body struct {
		Topology map[string][]string `json:"topology"`
	}

	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.topology = body.Topology

	response := map[string]any{
		"type": "topology_ok",
	}

	return s.node.Reply(msg, response)
}
