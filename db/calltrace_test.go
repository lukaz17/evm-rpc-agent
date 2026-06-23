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
	if firstTx.Result.From != "0xa30d8157911ef23c46c0eb71889efe6a648a41f7" {
		t.Errorf("first from = %s", firstTx.Result.From)
	}
	if firstTx.Result.To != "0xdef426319baf76cb4359e49268e05023b834f4df" {
		t.Errorf("first to = %s", firstTx.Result.To)
	}
	if firstTx.Result.Type != "CALL" {
		t.Errorf("first type = %s, want CALL", firstTx.Result.Type)
	}
	if firstTx.Result.Gas != "0xe57e0" {
		t.Errorf("first gas = %s, want 0xe57e0", firstTx.Result.Gas)
	}
	if firstTx.Result.GasUsed != "0x5208" {
		t.Errorf("first gasUsed = %s, want 0x5208", firstTx.Result.GasUsed)
	}
	if firstTx.Result.Value != "0x8510d6ff9c5b0000" {
		t.Errorf("first value = %s, want 0x8510d6ff9c5b0000", firstTx.Result.Value)
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
	if fetched.Data[0].Result.From != "0xa30d8157911ef23c46c0eb71889efe6a648a41f7" {
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

	coll := dbc.db.Collection(CallTracesCollection)
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

func TestInsertRawCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	raw := rawBlocks["0x4870d2"]

	if err := dbc.InsertRawCallTrace(ctx, 0x4870d2, raw); err != nil {
		t.Fatalf("InsertRawCallTrace: %v", err)
	}

	coll := dbc.db.Collection(CallTracesCollection)
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

	if err := dbc.InsertRawCallTrace(ctx, 0x4870d2, raw); err != nil {
		t.Fatalf("first InsertRawCallTrace: %v", err)
	}

	err := dbc.InsertRawCallTrace(ctx, 0x4870d2, raw)
	if err == nil {
		t.Error("expected duplicate key error")
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

func TestUpsertRawCallTrace(t *testing.T) {
	dbc := newTestCallTraceDbContext(t)
	ctx := context.Background()

	rawBlocks := loadJson(t, "blockCallTraces.json")
	raw1 := rawBlocks["0x4870d2"]

	if err := dbc.UpsertRawCallTrace(ctx, 0x4870d2, raw1); err != nil {
		t.Fatalf("first UpsertRawCallTrace: %v", err)
	}

	count, _ := dbc.CountCallTraces(ctx)
	if count != 1 {
		t.Errorf("count after first upsert = %d, want 1", count)
	}

	raw2 := rawBlocks["0x4870d2"]
	time.Sleep(10 * time.Millisecond)
	if err := dbc.UpsertRawCallTrace(ctx, 0x4870d2, raw2); err != nil {
		t.Fatalf("second UpsertRawCallTrace: %v", err)
	}

	count, _ = dbc.CountCallTraces(ctx)
	if count != 1 {
		t.Errorf("count after second upsert = %d, want 1 (upsert should not create duplicate)", count)
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

	coll := dbc.db.Collection(CallTracesCollection)
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
