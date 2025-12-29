package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	node := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(node)

	s := server{
		node: node,
		kv:   kv,
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
	kv   *maelstrom.KV
}

func (s *server) handleSend(msg maelstrom.Message) error {
	var body struct {
		Key string          `json:"key"`
		Msg json.RawMessage `json:"msg"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	offset, err := s.reserveOffset(context.Background(), body.Key)
	if err != nil {
		return err
	}

	msgCopy := json.RawMessage(append([]byte{}, body.Msg...))
	if err := s.kv.Write(context.Background(), messageKey(body.Key, offset), msgCopy); err != nil {
		return err
	}

	response := map[string]any{
		"type":   "send_ok",
		"offset": offset,
	}

	return s.node.Reply(msg, response)
}

func (s *server) handlePoll(msg maelstrom.Message) error {
	var body struct {
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	result := make(map[string][][]any, len(body.Offsets))

	for key, start := range body.Offsets {
		next, err := s.readNextOffset(context.Background(), key)
		if err != nil {
			return err
		}

		if start < 0 {
			start = 0
		}
		if start >= next {
			result[key] = [][]any{}
			continue
		}

		msgs := make([][]any, 0, next-start)
		for offset := start; offset < next; offset++ {
			val, err := s.kv.Read(context.Background(), messageKey(key, offset))
			if err != nil {
				if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
					break
				}
				return err
			}

			rawVal, err := json.Marshal(val)
			if err != nil {
				return err
			}

			msgs = append(msgs, []any{offset, json.RawMessage(rawVal)})
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

	for key, offset := range body.Offsets {
		if err := s.storeCommit(context.Background(), key, offset); err != nil {
			return err
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

	offsets := make(map[string]int, len(body.Keys))
	for _, key := range body.Keys {
		offset, err := s.kv.ReadInt(context.Background(), commitKey(key))
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				continue
			}
			return err
		}
		offsets[key] = offset
	}

	response := map[string]any{
		"type":    "list_committed_offsets_ok",
		"offsets": offsets,
	}

	return s.node.Reply(msg, response)
}

func (s *server) reserveOffset(ctx context.Context, key string) (int, error) {
	storageKey := nextOffsetKey(key)

	for {
		current, err := s.kv.ReadInt(ctx, storageKey)
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				current = 0
			} else {
				return 0, err
			}
		}

		if err := s.kv.CompareAndSwap(ctx, storageKey, current, current+1, true); err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.PreconditionFailed {
				continue
			}
			return 0, err
		}

		return current, nil
	}
}

func (s *server) readNextOffset(ctx context.Context, key string) (int, error) {
	val, err := s.kv.ReadInt(ctx, nextOffsetKey(key))
	if err != nil {
		if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}

func (s *server) storeCommit(ctx context.Context, key string, offset int) error {
	storageKey := commitKey(key)

	for {
		current, err := s.kv.ReadInt(ctx, storageKey)
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				current = -1
			} else {
				return err
			}
		}

		if offset <= current {
			return nil
		}

		if err := s.kv.CompareAndSwap(ctx, storageKey, current, offset, true); err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.PreconditionFailed {
				continue
			}
			return err
		}

		return nil
	}
}

func nextOffsetKey(key string) string {
	return fmt.Sprintf("log:%s:next", key)
}

func messageKey(key string, offset int) string {
	return fmt.Sprintf("log:%s:%d", key, offset)
}

func commitKey(key string) string {
	return fmt.Sprintf("commit:%s", key)
}
