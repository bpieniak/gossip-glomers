package main

import (
	"encoding/json"
	"fmt"
)

type TxnOperation struct {
	OperationType string
	Key           float64
	Value         *float64
}

func (op *TxnOperation) UnmarshalJSON(data []byte) error {
	var raw = []any{}

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	if len(raw) != 3 {
		return fmt.Errorf("expected 3 elements, got %d", len(raw))
	}

	var ok bool

	op.OperationType, ok = raw[0].(string)
	if !ok {
		return fmt.Errorf("invalid format for operation type '%s'", raw[0])
	}

	op.Key, ok = raw[1].(float64)
	if !ok {
		return fmt.Errorf("invalid format for operation key '%s'", raw[1])
	}

	if raw[2] == nil {
		op.Value = nil
	} else {
		opValue, ok := raw[2].(float64)
		if !ok {
			return fmt.Errorf("invalid format for operation value '%s'", raw[2])
		}

		op.Value = &opValue

	}

	return nil
}

func (op *TxnOperation) MarshalJSON() ([]byte, error) {
	out := []any{op.OperationType, op.Key, op.Value}
	return json.Marshal(out)
}
