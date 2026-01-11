package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	node := maelstrom.NewNode()

	s := server{
		node:  node,
		store: map[float64]float64{},
	}

	node.Handle("txn", s.handleTxn)
	node.Handle("replicate", s.handleReplicate)

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	node    *maelstrom.Node
	store   map[float64]float64
	storeMu sync.RWMutex
}

type TxnMsg struct {
	MsgID int            `json:"msg_id"`
	Txn   []TxnOperation `json:"txn"`
}

type ReplicateMsg struct {
	Txn []TxnOperation `json:"txn"`
}

func (s *server) handleTxn(msg maelstrom.Message) error {
	var txnMsg TxnMsg
	if err := json.Unmarshal(msg.Body, &txnMsg); err != nil {
		return err
	}

	responseTxns := make([]TxnOperation, 0, len(txnMsg.Txn))
	replicateTxns := make([]TxnOperation, 0, len(txnMsg.Txn))

	s.storeMu.Lock()

	for _, txn := range txnMsg.Txn {
		switch txn.OperationType {
		case "r":
			responseTxn := txn

			val, exists := s.store[txn.Key]
			if exists {
				responseTxn.Value = &val
			}

			responseTxns = append(responseTxns, responseTxn)
		case "w":
			if txn.Value == nil {
				s.storeMu.Unlock()
				return fmt.Errorf("missing value for write on key %v", txn.Key)
			}
			s.store[txn.Key] = *txn.Value

			responseTxns = append(responseTxns, txn)
			replicateTxns = append(replicateTxns, txn)
		default:
			s.storeMu.Unlock()
			return fmt.Errorf("invalid operation type '%s'", txn.OperationType)
		}

	}

	s.storeMu.Unlock()

	s.replicateWrites(replicateTxns)

	response := map[string]any{
		"type":        "txn_ok",
		"in_reply_to": txnMsg.MsgID,
		"txn":         responseTxns,
	}

	return s.node.Reply(msg, response)
}

func (s *server) handleReplicate(msg maelstrom.Message) error {
	var replicateMsg ReplicateMsg
	if err := json.Unmarshal(msg.Body, &replicateMsg); err != nil {
		return err
	}

	s.storeMu.Lock()
	for _, txn := range replicateMsg.Txn {
		if txn.OperationType != "w" || txn.Value == nil {
			continue
		}
		s.store[txn.Key] = *txn.Value
	}
	s.storeMu.Unlock()

	return nil
}

func (s *server) replicateWrites(txns []TxnOperation) {
	if len(txns) == 0 {
		return
	}

	msg := map[string]any{
		"type": "replicate",
		"txn":  txns,
	}

	for _, nodeID := range s.node.NodeIDs() {
		if nodeID == s.node.ID() {
			continue
		}

		go func(dest string) {
			if err := s.node.Send(dest, msg); err != nil {
				log.Printf("replicate to %s failed: %v", dest, err)
			}
		}(nodeID)
	}
}
