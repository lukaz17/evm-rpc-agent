// Copyright (C) 2025 Nguyen Nhat Tung
// This file is part of EVM RPC Agent
//
// EVM RPC Agent is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// EVM RPC Agent is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with EVM RPC Agent. If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/lukaz17/evm-rpc-agent/rpc"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func TestInsertAndGetCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

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
		t.Fatalf("NewBlockCallTraceFromRPC: %v", err)
	}

	if err := dbc.InsertCallTrace(ctx, trace); err != nil {
		t.Fatalf("InsertCallTrace: %v", err)
	}

	fetched, err := dbc.GetCallTrace(ctx, 0x4870d2)
	if err != nil {
		t.Fatalf("GetCallTrace: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected non-nil trace")
	}
	if fetched.ID != "4747474" {
		t.Errorf("ID = %s, want 4747474", fetched.ID)
	}
	if fetched.Number != 0x4870d2 {
		t.Errorf("Number = %d, want 0x4870d2", fetched.Number)
	}
	if len(fetched.Data) != 8 {
		t.Fatalf("Data count = %d, want 8", len(fetched.Data))
	}
	if fetched.Data[0].TxHash != "0xc237db67c16416d8f129cb1d5c36f801b2e3625fae7f29b98ca9e349e5752952" {
		t.Errorf("first txHash = %s", fetched.Data[0].TxHash)
	}
	if fetched.Data[0].Result.From.Hex() != "0xa30d8157911ef23c46c0eb71889efe6a648a41f7" {
		t.Errorf("first from = %s", fetched.Data[0].Result.From)
	}
}

func TestGetCallTrace_NonExistent(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	fetched, err := dbc.GetCallTrace(ctx, 999999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched != nil {
		t.Errorf("expected nil trace, got %+v", fetched)
	}
}

func TestGetRawCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

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
		t.Fatalf("NewBlockCallTraceFromRPC: %v", err)
	}

	if err := dbc.InsertCallTrace(ctx, trace); err != nil {
		t.Fatalf("InsertCallTrace: %v", err)
	}

	coll := dbc.db.Collection(CallTraceCollection)
	var doc bson.M
	err = coll.FindOne(ctx, bson.M{"_id": "4747474"}).Decode(&doc)
	if err != nil {
		t.Fatalf("find by _id: %v", err)
	}

	fetchedID := doc["_id"]
	if fetchedID != "4747474" {
		t.Errorf("_id = %v, want 4747474", fetchedID)
	}
}

func TestGetRawCallTrace_NonExistent(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	raw, err := dbc.GetRawCallTrace(ctx, 999999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw != nil {
		t.Errorf("expected nil document, got %+v", raw)
	}
}

func TestInsertCallTrace_Duplicate(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

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
		t.Fatalf("NewBlockCallTraceFromRPC: %v", err)
	}

	if err := dbc.InsertCallTrace(ctx, trace); err != nil {
		t.Fatalf("first InsertCallTrace: %v", err)
	}

	trace2, _ := NewBlockCallTraceFromRPC(0x4870d2, rpcTraces)
	err = dbc.InsertCallTrace(ctx, trace2)
	if err == nil {
		t.Error("expected duplicate key error")
	}
}

func TestInsertCallTraces(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	raw := loadJsonN(t, "blockCallTraces.json")
	var traces []*BlockCallTrace

	for _, key := range []string{"0x1a33b3", "0x1a33bb", "0x4870d2"} {
		tracesJSON := raw[key]
		var rpcTraces []rpc.TransactionTrace
		for _, tj := range tracesJSON {
			var tr rpc.TransactionTrace
			if err := json.Unmarshal(tj, &tr); err != nil {
				t.Fatalf("unmarshal trace: %v", err)
			}
			tr.Type = rpc.TraceTypeCall
			rpcTraces = append(rpcTraces, tr)
		}

		var blockNum uint64
		fmt.Sscanf(key, "0x%x", &blockNum)

		trace, err := NewBlockCallTraceFromRPC(blockNum, rpcTraces)
		if err != nil {
			t.Fatalf("NewBlockCallTraceFromRPC(%s): %v", key, err)
		}
		if trace != nil {
			traces = append(traces, trace)
		}
	}

	if err := dbc.InsertCallTraces(ctx, traces); err != nil {
		t.Fatalf("InsertCallTraces: %v", err)
	}

	count, _ := dbc.CountCallTraces(ctx)
	if count != int64(len(traces)) {
		t.Errorf("count = %d, want %d", count, len(traces))
	}

	for _, trace := range traces {
		fetched, err := dbc.GetCallTrace(ctx, trace.Number)
		if err != nil {
			t.Fatalf("GetCallTrace(%d): %v", trace.Number, err)
		}
		if fetched == nil {
			t.Errorf("expected non-nil trace for block %d", trace.Number)
		}
	}
}

func TestInsertCallTraces_Empty(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	err := dbc.InsertCallTraces(ctx, nil)
	if err != nil {
		t.Errorf("expected no error for nil traces, got %v", err)
	}

	err = dbc.InsertCallTraces(ctx, []*BlockCallTrace{})
	if err != nil {
		t.Errorf("expected no error for empty traces, got %v", err)
	}
}

func TestInsertCallTraces_RollbackOnDuplicate(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	raw := loadJsonN(t, "blockCallTraces.json")

	// Pre-insert trace A (0x4870d2).
	tracesJSONA := raw["0x4870d2"]
	var rpcTracesA []rpc.TransactionTrace
	for _, tj := range tracesJSONA {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace A: %v", err)
		}
		tr.Type = rpc.TraceTypeCall
		rpcTracesA = append(rpcTracesA, tr)
	}
	traceA, _ := NewBlockCallTraceFromRPC(0x4870d2, rpcTracesA)
	if err := dbc.InsertCallTrace(ctx, traceA); err != nil {
		t.Fatalf("InsertCallTrace A: %v", err)
	}

	// Prepare trace B (0x1a33bb) — new.
	tracesJSONB := raw["0x1a33bb"]
	var rpcTracesB []rpc.TransactionTrace
	for _, tj := range tracesJSONB {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace B: %v", err)
		}
		tr.Type = rpc.TraceTypeCall
		rpcTracesB = append(rpcTracesB, tr)
	}
	traceB, _ := NewBlockCallTraceFromRPC(0x1a33bb, rpcTracesB)

	err := dbc.InsertCallTraces(ctx, []*BlockCallTrace{traceB, traceA})
	if err == nil {
		t.Fatal("expected error from duplicate key")
	}

	count, err := dbc.CountCallTraces(ctx)
	if err != nil {
		t.Fatalf("CountCallTraces: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (B should have been rolled back)", count)
	}

	fetchedA, err := dbc.GetCallTrace(ctx, 0x4870d2)
	if err != nil || fetchedA == nil {
		t.Fatalf("GetCallTrace A: expected trace to exist, err=%v", err)
	}

	fetchedB, err := dbc.GetCallTrace(ctx, 0x1a33bb)
	if err != nil {
		t.Fatalf("GetCallTrace B: unexpected error %v", err)
	}
	if fetchedB != nil {
		t.Error("trace B should have been rolled back")
	}
}

func TestInsertCallTraces_RollbackMultipleInserted(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	raw := loadJsonN(t, "blockCallTraces.json")

	// Pre-insert trace A (0x4870d2).
	tracesJSONA := raw["0x4870d2"]
	var rpcTracesA []rpc.TransactionTrace
	for _, tj := range tracesJSONA {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace A: %v", err)
		}
		tr.Type = rpc.TraceTypeCall
		rpcTracesA = append(rpcTracesA, tr)
	}
	traceA, _ := NewBlockCallTraceFromRPC(0x4870d2, rpcTracesA)
	if err := dbc.InsertCallTrace(ctx, traceA); err != nil {
		t.Fatalf("InsertCallTrace A: %v", err)
	}

	// Prepare trace B (0x1a33bb) — new.
	tracesJSONB := raw["0x1a33bb"]
	var rpcTracesB []rpc.TransactionTrace
	for _, tj := range tracesJSONB {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace B: %v", err)
		}
		tr.Type = rpc.TraceTypeCall
		rpcTracesB = append(rpcTracesB, tr)
	}
	traceB, _ := NewBlockCallTraceFromRPC(0x1a33bb, rpcTracesB)

	// Craft a unique trace C with number 999998.
	traceC, _ := NewBlockCallTraceFromRPC(999998, rpcTracesB)
	traceC.Number = 999998
	traceC.ID = "999998"

	err := dbc.InsertCallTraces(ctx, []*BlockCallTrace{traceB, traceC, traceA})
	if err == nil {
		t.Fatal("expected error from duplicate key")
	}

	count, err := dbc.CountCallTraces(ctx)
	if err != nil {
		t.Fatalf("CountCallTraces: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (B and C should have been rolled back)", count)
	}

	fetchedC, err := dbc.GetCallTrace(ctx, 999998)
	if err != nil {
		t.Fatalf("GetCallTrace C: unexpected error %v", err)
	}
	if fetchedC != nil {
		t.Error("trace C should have been rolled back")
	}

	fetchedB, err := dbc.GetCallTrace(ctx, 0x1a33bb)
	if err != nil {
		t.Fatalf("GetCallTrace B: unexpected error %v", err)
	}
	if fetchedB != nil {
		t.Error("trace B should have been rolled back")
	}
}

func TestInsertRawCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	raw := rawBlocks["0x4870d2"]

	if err := dbc.InsertRawCallTrace(ctx, RawDataWithBlockNum{BlockNumber: 0x4870d2, Raw: raw}); err != nil {
		t.Fatalf("InsertRawCallTrace: %v", err)
	}

	coll := dbc.db.Collection(CallTraceCollection)
	var doc bson.M
	err := coll.FindOne(ctx, bson.M{"_id": "4747474"}).Decode(&doc)
	if err != nil {
		t.Fatalf("find by _id: %v", err)
	}

	num, ok := doc["number"]
	if !ok {
		t.Fatal("expected number field in raw call trace")
	}
	var numVal uint64
	switch v := num.(type) {
	case uint64:
		numVal = v
	case int64:
		numVal = uint64(v)
	case string:
		fmt.Sscanf(v, "0x%x", &numVal)
	default:
		t.Fatalf("unexpected number type %T", num)
	}
	if numVal != 0x4870d2 {
		t.Errorf("number = %v, want 0x4870d2", numVal)
	}
}

func TestInsertRawCallTrace_Duplicate(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	raw := rawBlocks["0x4870d2"]

	if err := dbc.InsertRawCallTrace(ctx, RawDataWithBlockNum{BlockNumber: 0x4870d2, Raw: raw}); err != nil {
		t.Fatalf("first InsertRawCallTrace: %v", err)
	}

	err := dbc.InsertRawCallTrace(ctx, RawDataWithBlockNum{BlockNumber: 0x4870d2, Raw: raw})
	if err == nil {
		t.Error("expected duplicate key error")
	}
}

func TestInsertRawCallTraces(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	keys := []string{"0x1a33b3", "0x1a33bb", "0x4870d2"}

	inputs := make([]RawDataWithBlockNum, 0, len(keys))
	for _, key := range keys {
		var blockNum uint64
		fmt.Sscanf(key, "0x%x", &blockNum)
		inputs = append(inputs, RawDataWithBlockNum{
			BlockNumber: blockNum,
			Raw:         rawBlocks[key],
		})
	}

	if err := dbc.InsertRawCallTraces(ctx, inputs); err != nil {
		t.Fatalf("InsertRawCallTraces: %v", err)
	}

	count, _ := dbc.CountCallTraces(ctx)
	if count != int64(len(inputs)) {
		t.Errorf("count = %d, want %d", count, len(inputs))
	}

	for _, input := range inputs {
		id := fmt.Sprintf("%d", input.BlockNumber)
		coll := dbc.db.Collection(CallTraceCollection)
		var doc bson.M
		err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
		if err != nil {
			t.Fatalf("FindOne(%s): %v", id, err)
		}
	}
}

func TestInsertRawCallTraces_Empty(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	err := dbc.InsertRawCallTraces(ctx, nil)
	if err != nil {
		t.Errorf("expected no error for nil inputs, got %v", err)
	}

	err = dbc.InsertRawCallTraces(ctx, []RawDataWithBlockNum{})
	if err != nil {
		t.Errorf("expected no error for empty inputs, got %v", err)
	}
}

func TestInsertRawCallTraces_RollbackOnDuplicate(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	rawA := rawBlocks["0x4870d2"]
	rawB := rawBlocks["0x1a33bb"]

	if err := dbc.InsertRawCallTrace(ctx, RawDataWithBlockNum{BlockNumber: 0x4870d2, Raw: rawA}); err != nil {
		t.Fatalf("InsertRawCallTrace A: %v", err)
	}

	err := dbc.InsertRawCallTraces(ctx, []RawDataWithBlockNum{
		{BlockNumber: 0x1a33bb, Raw: rawB},
		{BlockNumber: 0x4870d2, Raw: rawA},
	})
	if err == nil {
		t.Fatal("expected error from duplicate key")
	}

	count, err := dbc.CountCallTraces(ctx)
	if err != nil {
		t.Fatalf("CountCallTraces: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (B should have been rolled back)", count)
	}

	fetchedA, err := dbc.GetCallTrace(ctx, 0x4870d2)
	if err != nil || fetchedA == nil {
		t.Fatalf("GetCallTrace A: expected trace to exist, err=%v", err)
	}

	fetchedB, err := dbc.GetCallTrace(ctx, 0x1a33bb)
	if err != nil {
		t.Fatalf("GetCallTrace B: unexpected error %v", err)
	}
	if fetchedB != nil {
		t.Error("trace B should have been rolled back")
	}
}

func TestInsertRawCallTraces_RollbackMultipleInserted(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	rawA := rawBlocks["0x4870d2"]
	rawB := rawBlocks["0x1a33bb"]

	if err := dbc.InsertRawCallTrace(ctx, RawDataWithBlockNum{BlockNumber: 0x4870d2, Raw: rawA}); err != nil {
		t.Fatalf("InsertRawCallTrace A: %v", err)
	}

	err := dbc.InsertRawCallTraces(ctx, []RawDataWithBlockNum{
		{BlockNumber: 0x1a33bb, Raw: rawB},
		{BlockNumber: 0xf4246, Raw: rawB},
		{BlockNumber: 0x4870d2, Raw: rawA},
	})
	if err == nil {
		t.Fatal("expected error from duplicate key")
	}

	count, err := dbc.CountCallTraces(ctx)
	if err != nil {
		t.Fatalf("CountCallTraces: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (B and C should have been rolled back)", count)
	}

	fetchedC, err := dbc.GetCallTrace(ctx, 0xf4246)
	if err != nil {
		t.Fatalf("GetCallTrace C: unexpected error %v", err)
	}
	if fetchedC != nil {
		t.Error("trace C should have been rolled back")
	}

	fetchedB, err := dbc.GetCallTrace(ctx, 0x1a33bb)
	if err != nil {
		t.Fatalf("GetCallTrace B: unexpected error %v", err)
	}
	if fetchedB != nil {
		t.Error("trace B should have been rolled back")
	}
}

func TestUpsertCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

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
		t.Fatalf("NewBlockCallTraceFromRPC: %v", err)
	}

	if err := dbc.UpsertCallTrace(ctx, trace); err != nil {
		t.Fatalf("UpsertCallTrace first: %v", err)
	}

	trace.Data[0].TxHash = "0xmodified"
	time.Sleep(10 * time.Millisecond)
	if err := dbc.UpsertCallTrace(ctx, trace); err != nil {
		t.Fatalf("UpsertCallTrace second: %v", err)
	}

	fetched, err := dbc.GetCallTrace(ctx, 0x4870d2)
	if err != nil {
		t.Fatalf("GetCallTrace: %v", err)
	}
	if fetched.Data[0].TxHash != "0xmodified" {
		t.Errorf("txHash = %s, want 0xmodified", fetched.Data[0].TxHash)
	}

	count, err := dbc.CountCallTraces(ctx)
	if err != nil {
		t.Fatalf("CountCallTraces: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (upsert should not create duplicate)", count)
	}
}

func TestUpsertCallTraces(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	raw := loadJsonN(t, "blockCallTraces.json")
	var traces []*BlockCallTrace

	for _, key := range []string{"0x1a33b3", "0x4870d2"} {
		tracesJSON := raw[key]
		var rpcTraces []rpc.TransactionTrace
		for _, tj := range tracesJSON {
			var tr rpc.TransactionTrace
			if err := json.Unmarshal(tj, &tr); err != nil {
				t.Fatalf("unmarshal trace: %v", err)
			}
			tr.Type = rpc.TraceTypeCall
			rpcTraces = append(rpcTraces, tr)
		}

		var blockNum uint64
		fmt.Sscanf(key, "0x%x", &blockNum)

		trace, err := NewBlockCallTraceFromRPC(blockNum, rpcTraces)
		if err != nil {
			t.Fatalf("NewBlockCallTraceFromRPC(%s): %v", key, err)
		}
		if trace != nil {
			traces = append(traces, trace)
		}
	}

	if err := dbc.UpsertCallTraces(ctx, traces); err != nil {
		t.Fatalf("first UpsertCallTraces: %v", err)
	}

	count, _ := dbc.CountCallTraces(ctx)
	if count != int64(len(traces)) {
		t.Errorf("count after first upsert = %d, want %d", count, len(traces))
	}

	for _, trace := range traces {
		trace.Data[0].TxHash = "0xmodified"
	}

	time.Sleep(10 * time.Millisecond)

	if err := dbc.UpsertCallTraces(ctx, traces); err != nil {
		t.Fatalf("second UpsertCallTraces: %v", err)
	}

	count, _ = dbc.CountCallTraces(ctx)
	if count != int64(len(traces)) {
		t.Errorf("count after second upsert = %d, want %d (upsert should not create duplicates)", count, len(traces))
	}

	for _, trace := range traces {
		fetched, err := dbc.GetCallTrace(ctx, trace.Number)
		if err != nil {
			t.Fatalf("GetCallTrace(%d): %v", trace.Number, err)
		}
		if fetched.Data[0].TxHash != "0xmodified" {
			t.Errorf("block %d txHash = %s, want 0xmodified", trace.Number, fetched.Data[0].TxHash)
		}
	}
}

func TestUpsertCallTraces_Empty(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	err := dbc.UpsertCallTraces(ctx, nil)
	if err != nil {
		t.Errorf("expected no error for nil traces, got %v", err)
	}

	err = dbc.UpsertCallTraces(ctx, []*BlockCallTrace{})
	if err != nil {
		t.Errorf("expected no error for empty traces, got %v", err)
	}
}

func TestUpsertRawCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	raw1 := rawBlocks["0x4870d2"]

	if err := dbc.UpsertRawCallTrace(ctx, RawDataWithBlockNum{BlockNumber: 0x4870d2, Raw: raw1}); err != nil {
		t.Fatalf("first UpsertRawCallTrace: %v", err)
	}

	count, _ := dbc.CountCallTraces(ctx)
	if count != 1 {
		t.Errorf("count after first upsert = %d, want 1", count)
	}

	raw2 := rawBlocks["0x4870d2"]
	time.Sleep(10 * time.Millisecond)
	if err := dbc.UpsertRawCallTrace(ctx, RawDataWithBlockNum{BlockNumber: 0x4870d2, Raw: raw2}); err != nil {
		t.Fatalf("second UpsertRawCallTrace: %v", err)
	}

	count, _ = dbc.CountCallTraces(ctx)
	if count != 1 {
		t.Errorf("count after second upsert = %d, want 1 (upsert should not create duplicate)", count)
	}
}

func TestUpsertRawCallTraces(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	keys := []string{"0x1a33b3", "0x4870d2"}

	inputs := make([]RawDataWithBlockNum, 0, len(keys))
	for _, key := range keys {
		var blockNum uint64
		fmt.Sscanf(key, "0x%x", &blockNum)
		inputs = append(inputs, RawDataWithBlockNum{
			BlockNumber: blockNum,
			Raw:         rawBlocks[key],
		})
	}

	if err := dbc.UpsertRawCallTraces(ctx, inputs); err != nil {
		t.Fatalf("first UpsertRawCallTraces: %v", err)
	}

	count, _ := dbc.CountCallTraces(ctx)
	if count != int64(len(inputs)) {
		t.Errorf("count after first upsert = %d, want %d", count, len(inputs))
	}

	time.Sleep(10 * time.Millisecond)

	if err := dbc.UpsertRawCallTraces(ctx, inputs); err != nil {
		t.Fatalf("second UpsertRawCallTraces: %v", err)
	}

	count, _ = dbc.CountCallTraces(ctx)
	if count != int64(len(inputs)) {
		t.Errorf("count after second upsert = %d, want %d (upsert should not create duplicates)", count, len(inputs))
	}
}

func TestUpsertRawCallTraces_Empty(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	err := dbc.UpsertRawCallTraces(ctx, nil)
	if err != nil {
		t.Errorf("expected no error for nil inputs, got %v", err)
	}

	err = dbc.UpsertRawCallTraces(ctx, []RawDataWithBlockNum{})
	if err != nil {
		t.Errorf("expected no error for empty inputs, got %v", err)
	}
}

func TestDeleteCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

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
		t.Fatalf("NewBlockCallTraceFromRPC: %v", err)
	}

	if err := dbc.InsertCallTrace(ctx, trace); err != nil {
		t.Fatalf("InsertCallTrace: %v", err)
	}

	countBefore, _ := dbc.CountCallTraces(ctx)
	if countBefore != 1 {
		t.Errorf("count before delete = %d, want 1", countBefore)
	}

	if err := dbc.DeleteCallTrace(ctx, 0x4870d2); err != nil {
		t.Fatalf("DeleteCallTrace: %v", err)
	}

	fetched, err := dbc.GetCallTrace(ctx, 0x4870d2)
	if err != nil {
		t.Fatalf("GetCallTrace after delete: %v", err)
	}
	if fetched != nil {
		t.Error("expected nil trace after delete")
	}

	countAfter, _ := dbc.CountCallTraces(ctx)
	if countAfter != 0 {
		t.Errorf("count after delete = %d, want 0", countAfter)
	}
}

func TestDeleteCallTrace_NonExistent(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	err := dbc.DeleteCallTrace(ctx, 999999)
	if err != nil {
		t.Errorf("unexpected error deleting non-existent trace: %v", err)
	}
}

func TestCountCallTraces(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	count, err := dbc.CountCallTraces(ctx)
	if err != nil {
		t.Fatalf("CountCallTraces: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	raw1 := loadJsonN(t, "blockCallTraces.json")
	tracesJSON1 := raw1["0x4870d2"]

	var rpcTraces1 []rpc.TransactionTrace
	for _, tj := range tracesJSON1 {
		var tr rpc.TransactionTrace
		if err := json.Unmarshal(tj, &tr); err != nil {
			t.Fatalf("unmarshal trace: %v", err)
		}
		tr.Type = rpc.TraceTypeCall
		rpcTraces1 = append(rpcTraces1, tr)
	}

	trace1, _ := NewBlockCallTraceFromRPC(0x4870d2, rpcTraces1)
	if err := dbc.InsertCallTrace(ctx, trace1); err != nil {
		t.Fatalf("InsertCallTrace 1: %v", err)
	}

	raw2 := loadJsonN(t, "blockCallTraces.json")
	tracesJSON2, ok2 := raw2["0x4870d3"]
	var rpcTraces2 []rpc.TransactionTrace
	if ok2 {
		for _, tj := range tracesJSON2 {
			var tr rpc.TransactionTrace
			if err := json.Unmarshal(tj, &tr); err != nil {
				t.Fatalf("unmarshal trace: %v", err)
			}
			tr.Type = rpc.TraceTypeCall
			rpcTraces2 = append(rpcTraces2, tr)
		}
	}

	trace2, _ := NewBlockCallTraceFromRPC(0x4870d3, rpcTraces2)
	if trace2 != nil {
		if err := dbc.InsertCallTrace(ctx, trace2); err != nil {
			t.Fatalf("InsertCallTrace 2: %v", err)
		}
	}

	count, err = dbc.CountCallTraces(ctx)
	if err != nil {
		t.Fatalf("CountCallTraces after insert: %v", err)
	}
	expected := int64(1)
	if trace2 != nil {
		expected = 2
	}
	if count != expected {
		t.Errorf("count = %d, want %d", count, expected)
	}
}

func newTestCallTraceDbContext(t *testing.T) *DbContext {
	t.Helper()
	dbc := &DbContext{
		uri:    dbUri,
		dbName: dbName,
	}
	if err := dbc.Connect(); err != nil {
		t.Fatalf("connect to mongodb: %v", err)
	}

	coll := dbc.db.Collection(CallTraceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := coll.DeleteMany(ctx, bson.M{}); err != nil {
		t.Fatalf("clean collection: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := coll.DeleteMany(ctx, bson.M{}); err != nil {
			t.Logf("cleanup failed: %v", err)
		}
		if err := dbc.Disconnect(); err != nil {
			t.Logf("disconnect failed: %v", err)
		}
	})

	return dbc
}
