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

func (s *server) handleTxn(msg maelstrom.Message) error {
	var txnMsg TxnMsg
	if err := json.Unmarshal(msg.Body, &txnMsg); err != nil {
		return err
	}

	responseTxns := make([]TxnOperation, 0, len(txnMsg.Txn))

	s.storeMu.Lock()
	defer s.storeMu.Unlock()

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
				return fmt.Errorf("missing value for write on key %v", txn.Key)
			}
			s.store[txn.Key] = *txn.Value

			responseTxns = append(responseTxns, txn)
		default:
			return fmt.Errorf("invalid operation type '%s'", txn.OperationType)
		}

	}

	response := map[string]any{
		"type":        "txn_ok",
		"in_reply_to": txnMsg.MsgID,
		"txn":         responseTxns,
	}

	return s.node.Reply(msg, response)
}
