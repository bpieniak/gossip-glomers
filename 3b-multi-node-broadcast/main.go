package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"golang.org/x/sync/errgroup"
)

func main() {
	n := maelstrom.NewNode()

	s := server{
		node: n,
	}

	n.Handle("broadcast", s.handleBroadcast)
	n.Handle("read", s.handleRead)
	n.Handle("topology", s.handleTopology)
	n.Handle("broadcast_ok", func(msg maelstrom.Message) error { return nil })

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	node *maelstrom.Node

	topology map[string][]string

	valuesMu sync.Mutex
	values   []float64
}

func (s *server) handleBroadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	message := body["message"].(float64)

	s.valuesMu.Lock()
	s.values = append(s.values, message)
	s.valuesMu.Unlock()

	err := s.broadcastToNeighbors(message, msg.Src)
	if err != nil {
		return err
	}

	response := map[string]any{
		"type": "broadcast_ok",
	}

	return s.node.Reply(msg, response)
}

func (s *server) broadcastToNeighbors(id float64, src string) error {
	errGroup := errgroup.Group{}

	for _, nodeId := range s.topology[s.node.ID()] {
		errGroup.Go(func() error {
			if nodeId == src {
				return nil
			}

			msg := map[string]any{
				"type":    "broadcast",
				"message": id,
			}

			return s.node.Send(nodeId, msg)
		})
	}

	return errGroup.Wait()
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
