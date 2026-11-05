// Copyright (C) 2025  Nguyen Nhat Tung
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"encoding/json"
	"testing"

	"github.com/lukaz17/evm-rpc-agent/rpc"
)

func TestNewBlockCallTraceFromRPC_Empty(t *testing.T) {
	result, err := NewBlockCallTraceFromRPC(0, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil trace, got %+v", result)
	}
}

func TestNewBlockCallTraceFromRPC_Simple(t *testing.T) {
	raw := loadJsonN(t, "blockCallTraces.json")
	tracesJSON := raw["0x4870d2"]

	var rpcTraces []rpc.TransactionTrace
	for _, tj := range tracesJSON {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace: %v", err)
		}
		tr.Type = rpc.TraceTypeCall
		rpcTraces = append(rpcTraces, tr)
	}

	blockNum := uint64(0x4870d2)

	trace, err := NewBlockCallTraceFromRPC(blockNum, rpcTraces)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trace.ID != "4747474" {
		t.Errorf("ID = %s, want 4747474", trace.ID)
	}
	if trace.Number != blockNum {
		t.Errorf("Number = %d (0x%x), want %d (0x4870d2)", trace.Number, trace.Number, blockNum)
	}
	if len(trace.Data) != 8 {
		t.Fatalf("Data count = %d, want 8", len(trace.Data))
	}
	if trace.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if trace.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	firstTx := trace.Data[0]
	if firstTx.TxHash != "0xc237db67c16416d8f129cb1d5c36f801b2e3625fae7f29b98ca9e349e5752952" {
		t.Errorf("first txHash = %s", firstTx.TxHash)
	}
	if firstTx.Result.From.Hex() != "0xa30d8157911ef23c46c0eb71889efe6a648a41f7" {
		t.Errorf("first from = %s", firstTx.Result.From)
	}
	if firstTx.Result.To.Hex() != "0xdef426319baf76cb4359e49268e05023b834f4df" {
		t.Errorf("first to = %s", firstTx.Result.To)
	}
	if firstTx.Result.Type != "CALL" {
		t.Errorf("first type = %s, want CALL", firstTx.Result.Type)
	}
	if firstTx.Result.Gas.Hex() != "0xe57e0" {
		t.Errorf("first gas = %s, want 0xe57e0", firstTx.Result.Gas.Hex())
	}
	if firstTx.Result.GasUsed.Hex() != "0x5208" {
		t.Errorf("first gasUsed = %s, want 0x5208", firstTx.Result.GasUsed.Hex())
	}
	if firstTx.Result.Value.Hex() != "0x8510d6ff9c5b0000" {
		t.Errorf("first value = %s, want 0x8510d6ff9c5b0000", firstTx.Result.Value.Hex())
	}
	if len(firstTx.Result.Calls) != 0 {
		t.Errorf("expected no sub-calls, got %d", len(firstTx.Result.Calls))
	}
}

func TestNewBlockCallTraceFromRPC_NestedCalls(t *testing.T) {
	raw := loadJsonN(t, "blockCallTraces.json")
	tracesJSON := raw["0x4870d2"]

	var rpcTraces []rpc.TransactionTrace
	for _, tj := range tracesJSON {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace: %v", err)
		}
		tr.Type = rpc.TraceTypeCall
		rpcTraces = append(rpcTraces, tr)
	}

	trace, err := NewBlockCallTraceFromRPC(0x4870d2, rpcTraces)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nestedTx := trace.Data[6]
	cf := nestedTx.Result

	if len(cf.Calls) == 0 {
		t.Fatal("expected sub-calls")
	}
	if cf.Calls[0].Type != "CALL" {
		t.Errorf("first sub-call type = %s, want CALL", cf.Calls[0].Type)
	}
	if len(cf.Calls[1].Calls) == 0 {
		t.Fatal("expected nested sub-calls at depth 2")
	}
	if cf.Calls[1].Calls[0].Type != "CALL" {
		t.Errorf("nested sub-call type = %s, want CALL", cf.Calls[1].Calls[0].Type)
	}
}

func TestNewBlockCallTraceFromRPC_WrongType(t *testing.T) {
	raw := loadJsonN(t, "blockTraces.json")
	tracesJSON := raw["0x1a33b7"]

	var rpcTraces []rpc.TransactionTrace
	for _, tj := range tracesJSON {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace: %v", err)
		}
		tr.Type = rpc.TraceTypeDefault
		rpcTraces = append(rpcTraces, tr)
	}

	_, err := NewBlockCallTraceFromRPC(0x1a33b7, rpcTraces)
	if err == nil {
		t.Error("expected error for non-call trace type")
	}
}
