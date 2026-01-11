package main_test

import (
	"bytes"
	"testing"

	main "github.com/bpieniak/gossip-glomers/6a-single-node-totally-available-transactions"
)

func TestTxnOperation_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		data    []byte
		wantErr bool
	}{
		{
			data:    []byte(`["r", 1, null]`),
			wantErr: false,
		},

		{
			data:    []byte(`["w", 1, 6]`),
			wantErr: false,
		},
		{
			data:    []byte(`["w", 2, 9]`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run("unmarshall", func(t *testing.T) {
			var op main.TxnOperation
			gotErr := op.UnmarshalJSON(tt.data)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalJSON() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UnmarshalJSON() succeeded unexpectedly")
			}
		})
	}
}

func TestTxnOperation_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		op      main.TxnOperation
		want    []byte
		wantErr bool
	}{
		{
			name: "read operation",
			op: main.TxnOperation{
				OperationType: "r",
				Key:           1,
			},
			want:    []byte(`["r",1,null]`),
			wantErr: false,
		},
		{
			name: "write operation",
			op: func() main.TxnOperation {
				val := float64(9)
				return main.TxnOperation{
					OperationType: "w",
					Key:           2,
					Value:         &val,
				}
			}(),
			want:    []byte(`["w",2,9]`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := tt.op
			got, gotErr := op.MarshalJSON()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MarshallJSON() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MarshallJSON() succeeded unexpectedly")
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("MarshallJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
