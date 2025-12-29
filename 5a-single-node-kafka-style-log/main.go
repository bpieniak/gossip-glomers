package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	node := maelstrom.NewNode()

	s := server{
		node: node,
	}

	node.Handle("send", s.handleSend)
	node.Handle("poll", s.handlePoll)
	node.Handle("commit_offsets", s.handleCommitOffsets)
	node.Handle("list_committed_offsets", s.handleListCommittedOffsets)

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	node *maelstrom.Node

	mu   sync.Mutex
	logs map[string]*logState
}

type logState struct {
	messages  []logEntry
	committed int
}

type logEntry struct {
	offset int
	value  json.RawMessage
}

func (s *server) handleSend(msg maelstrom.Message) error {
	var body struct {
		Key string          `json:"key"`
		Msg json.RawMessage `json:"msg"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	msgCopy := json.RawMessage(append([]byte{}, body.Msg...))

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logs == nil {
		s.logs = make(map[string]*logState)
	}

	state, ok := s.logs[body.Key]
	if !ok {
		state = &logState{committed: -1}
		s.logs[body.Key] = state
	}

	offset := len(state.messages)
	state.messages = append(state.messages, logEntry{
		offset: offset,
		value:  msgCopy,
	})

	response := map[string]any{
		"type":   "send_ok",
		"offset": offset,
	}

	return s.node.Reply(msg, response)
}

const maxPollCount = 5

func (s *server) handlePoll(msg maelstrom.Message) error {
	var body struct {
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string][][]any, len(body.Offsets))

	for key, start := range body.Offsets {
		state, ok := s.logs[key]
		if !ok {
			result[key] = [][]any{}
			continue
		}

		if start < 0 {
			start = 0
		}
		if start >= len(state.messages) {
			result[key] = [][]any{}
			continue
		}

		msgs := make([][]any, 0, len(state.messages)-start)
		for i := start; i < len(state.messages) && i <= start+maxPollCount; i++ {
			entry := state.messages[i]
			msgs = append(msgs, []any{entry.offset, entry.value})
		}

		result[key] = msgs
	}

	response := map[string]any{
		"type": "poll_ok",
		"msgs": result,
	}

	return s.node.Reply(msg, response)
}

func (s *server) handleCommitOffsets(msg maelstrom.Message) error {
	var body struct {
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logs == nil {
		s.logs = make(map[string]*logState)
	}

	for key, offset := range body.Offsets {
		state, ok := s.logs[key]
		if !ok {
			state = &logState{committed: -1}
			s.logs[key] = state
		}

		if offset > state.committed {
			state.committed = offset
		}
	}

	response := map[string]any{
		"type": "commit_offsets_ok",
	}

	return s.node.Reply(msg, response)
}

func (s *server) handleListCommittedOffsets(msg maelstrom.Message) error {
	var body struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	offsets := make(map[string]int, len(body.Keys))
	for _, key := range body.Keys {
		if state, ok := s.logs[key]; ok && state.committed >= 0 {
			offsets[key] = state.committed
		}
	}

	response := map[string]any{
		"type":    "list_committed_offsets_ok",
		"offsets": offsets,
	}

	return s.node.Reply(msg, response)
}
