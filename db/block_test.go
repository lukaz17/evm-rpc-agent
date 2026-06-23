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

const (
	dbUri  = "mongodb://localhost:27017"
	dbName = "evm-rpc-agent-test"
)

func TestNewBlockFromRPC_Nil(t *testing.T) {
	result, err := NewBlockFromRPC(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil block, got %+v", result)
	}
}

func TestNewBlockFromRPC_EmptyTX(t *testing.T) {
	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if block.ID != "272727" {
		t.Errorf("ID = %s, want 272727", block.ID)
	}
	if block.Number != 0x42957 {
		t.Errorf("Number = %d (0x%x), want 272727 (0x42957)", block.Number, block.Number)
	}
	if block.Data.Hash != "0x58e72ae0b02c13575f9124ed304cea92c1cfc8c9f8ed905d3bb2b57028a1e123" {
		t.Errorf("Hash = %s", block.Data.Hash)
	}
	if block.Data.Miner != "0xf8b483dba2c3b7176a3da549ad41a48bb3121069" {
		t.Errorf("Miner = %s", block.Data.Miner)
	}
	if block.Data.GasUsed != 0 {
		t.Errorf("GasUsed = %d, want 0", block.Data.GasUsed)
	}
	if block.Data.GasLimit != 0x2fefd8 {
		t.Errorf("GasLimit = 0x%x, want 0x2fefd8", block.Data.GasLimit)
	}
	if len(block.Data.Transactions) != 0 {
		t.Errorf("Transactions count = %d, want 0", len(block.Data.Transactions))
	}
	if len(block.Data.Uncles) != 0 {
		t.Errorf("Uncles count = %d, want 0", len(block.Data.Uncles))
	}
	if block.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if block.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestNewBlockFromRPC_WithTX(t *testing.T) {
	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x1a33b7"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x1a33b7: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if block.ID != "1717175" {
		t.Errorf("ID = %s, want 1717175", block.ID)
	}
	if block.Number != 0x1a33b7 {
		t.Errorf("Number = %d (0x%x), want 1717175 (0x1a33b7)", block.Number, block.Number)
	}
	if block.Data.Hash != "0x463b405b6b89683dd05956aaed190b3ab63c2d0efdbf1a99c5cd9f5aaeab11fc" {
		t.Errorf("Hash = %s", block.Data.Hash)
	}
	if block.Data.GasUsed != 0x63003 {
		t.Errorf("GasUsed = 0x%x, want 0x63003", block.Data.GasUsed)
	}

	if len(block.Data.Transactions) != 9 {
		t.Fatalf("Transactions count = %d, want 9", len(block.Data.Transactions))
	}

	firstTx := block.Data.Transactions[0]
	if firstTx.Hash != "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686" {
		t.Errorf("first tx hash = %s", firstTx.Hash)
	}
	if firstTx.From != "0xfbe26da0e985087d28228defbdaa394713b0865f" {
		t.Errorf("first tx from = %s", firstTx.From)
	}
	if firstTx.To != "0xfb110ca742c3e47ab3babf4d48624c33891f35c6" {
		t.Errorf("first tx to = %v", firstTx.To)
	}
	if firstTx.Gas != 0x5208 {
		t.Errorf("first tx gas = 0x%x, want 0x5208", firstTx.Gas)
	}

	lastTx := block.Data.Transactions[8]
	if lastTx.Input != "0x667a2f58" {
		t.Errorf("last tx input = %s, want 0x667a2f58", lastTx.Input)
	}
}

func TestNewTransactionFromRPC(t *testing.T) {
	txJSON := `{
		"blockHash": "0x463b405b6b89683dd05956aaed190b3ab63c2d0efdbf1a99c5cd9f5aaeab11fc",
		"blockNumber": "0x1a33b7",
		"blockTimestamp": "0x57632695",
		"from": "0xfbe26da0e985087d28228defbdaa394713b0865f",
		"gas": "0x5208",
		"gasPrice": "0x9502f9000",
		"hash": "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686",
		"input": "0x",
		"nonce": "0x1390",
		"to": "0xfb110ca742c3e47ab3babf4d48624c33891f35c6",
		"transactionIndex": "0x0",
		"value": "0x1062fab684fee0",
		"type": "0x0",
		"v": "0x1c",
		"r": "0x4b089b013cad39edd3fb1ee8dbe36e131a570aa807baea80efb8c61b2967d396",
		"s": "0x6d2ef2f6135f38c61ca13b340265fc587c2dcec980c24294807b105279b4efca"
	}`

	var rpcTx rpc.Transaction
	if err := json.Unmarshal([]byte(txJSON), &rpcTx); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	tx := NewTransactionFromRPC(rpcTx)

	if tx.Hash != "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686" {
		t.Errorf("hash = %s", tx.Hash)
	}
	if tx.BlockNumber != 0x1a33b7 {
		t.Errorf("blockNumber = %d, want 0x1a33b7", tx.BlockNumber)
	}
	if tx.From != "0xfbe26da0e985087d28228defbdaa394713b0865f" {
		t.Errorf("from = %s", tx.From)
	}
	if tx.To != "0xfb110ca742c3e47ab3babf4d48624c33891f35c6" {
		t.Errorf("to = %v", tx.To)
	}
	if tx.Gas != 0x5208 {
		t.Errorf("gas = 0x%x, want 0x5208", tx.Gas)
	}
	if tx.Index != 0 {
		t.Errorf("index = %d, want 0", tx.Index)
	}
}

func TestInsertAndGetBlock(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("NewBlockFromRPC: %v", err)
	}

	if err := dbc.InsertBlock(ctx, block); err != nil {
		t.Fatalf("InsertBlock: %v", err)
	}

	fetched, err := dbc.GetBlock(ctx, 0x42957)
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected non-nil block")
	}
	if fetched.ID != "272727" {
		t.Errorf("ID = %s, want 272727", fetched.ID)
	}
	if fetched.Number != 0x42957 {
		t.Errorf("Number = %d, want 0x42957", fetched.Number)
	}
	if fetched.Data.Hash != block.Data.Hash {
		t.Errorf("Hash = %s, want %s", fetched.Data.Hash, block.Data.Hash)
	}
}

func TestGetBlock_NonExistent(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	fetched, err := dbc.GetBlock(ctx, 999999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched != nil {
		t.Errorf("expected nil block, got %+v", fetched)
	}
}

func TestGetRawBlock(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("NewBlockFromRPC: %v", err)
	}

	if err := dbc.InsertBlock(ctx, block); err != nil {
		t.Fatalf("InsertBlock: %v", err)
	}

	raw, err := dbc.GetRawBlock(ctx, 272727)
	if err != nil {
		t.Fatalf("GetRawBlock: %v", err)
	}
	if raw == nil {
		t.Fatal("expected non-nil document")
	}

	if raw.Number != 272727 {
		t.Errorf("number = %v, want 0x42957", raw.Number)
	}
}

func TestGetRawBlock_NonExistent(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	raw, err := dbc.GetRawBlock(ctx, 999999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw != nil {
		t.Errorf("expected nil document, got %+v", raw)
	}
}

func TestInsertAndGetBlock_WithTransactions(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x1a33b7"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x1a33b7: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("NewBlockFromRPC: %v", err)
	}

	if err := dbc.InsertBlock(ctx, block); err != nil {
		t.Fatalf("InsertBlock: %v", err)
	}

	fetched, err := dbc.GetBlock(ctx, 1717175)
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected non-nil block")
	}
	if len(fetched.Data.Transactions) != 9 {
		t.Errorf("Transactions count = %d, want 9", len(fetched.Data.Transactions))
	}
	if fetched.Data.Transactions[0].Hash != "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686" {
		t.Errorf("first tx hash = %s", fetched.Data.Transactions[0].Hash)
	}
}

func TestInsertAndGetBlock_BlockNumberAsID(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("NewBlockFromRPC: %v", err)
	}

	if err := dbc.InsertBlock(ctx, block); err != nil {
		t.Fatalf("InsertBlock: %v", err)
	}

	coll := dbc.db.Collection(BlockCollection)
	var doc bson.M
	err = coll.FindOne(ctx, bson.M{"_id": "272727"}).Decode(&doc)
	if err != nil {
		t.Fatalf("direct find by block number ID: %v", err)
	}

	fetchedID := doc["_id"]
	if fetchedID != "272727" {
		t.Errorf("_id = %v, want 272727 (block number as string)", fetchedID)
	}

	err = coll.FindOne(ctx, bson.M{"_id": "block:272727"}).Decode(&doc)
	if err == nil {
		t.Error("should not find document with block: prefix ID")
	}
}

func TestInsertBlock_Duplicate(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	var raw rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &raw); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block, err := NewBlockFromRPC(&raw)
	if err != nil {
		t.Fatalf("NewBlockFromRPC: %v", err)
	}

	if err := dbc.InsertBlock(ctx, block); err != nil {
		t.Fatalf("first InsertBlock: %v", err)
	}

	block2, _ := NewBlockFromRPC(&raw)
	err = dbc.InsertBlock(ctx, block2)
	if err == nil {
		t.Error("expected duplicate key error")
	}
}

func TestInsertRawBlock(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	raws := loadJson(t, "blocks.json")
	raw := raws["0x42957"]

	if err := dbc.InsertRawBlock(ctx, raw); err != nil {
		t.Fatalf("InsertRawBlock: %v", err)
	}

	coll := dbc.db.Collection(BlockCollection)
	var doc bson.M
	err := coll.FindOne(ctx, bson.M{"_id": "272727"}).Decode(&doc)
	if err != nil {
		t.Fatalf("find by _id 0 (hex number stored as string): %v", err)
	}

	num, ok := doc["number"]
	if !ok {
		t.Fatal("expected number field in raw block")
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
	if numVal != 272727 {
		t.Errorf("number = %v, want 272727", numVal)
	}
}

func TestInsertRawBlock_Duplicate(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	raws := loadJson(t, "blocks.json")
	raw := raws["0x42957"]

	if err := dbc.InsertRawBlock(ctx, raw); err != nil {
		t.Fatalf("first InsertRawBlock: %v", err)
	}

	err := dbc.InsertRawBlock(ctx, raw)
	if err == nil {
		t.Error("expected duplicate key error")
	}
}

func TestUpsertBlock(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("NewBlockFromRPC: %v", err)
	}

	if err := dbc.UpsertBlock(ctx, block); err != nil {
		t.Fatalf("UpsertBlock first: %v", err)
	}

	block.Data.Hash = "0xmodified"
	time.Sleep(10 * time.Millisecond)
	if err := dbc.UpsertBlock(ctx, block); err != nil {
		t.Fatalf("UpsertBlock second: %v", err)
	}

	fetched, err := dbc.GetBlock(ctx, 0x42957)
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if fetched.Data.Hash != "0xmodified" {
		t.Errorf("Hash = %s, want 0xmodified", fetched.Data.Hash)
	}

	count, err := dbc.CountBlocks(ctx)
	if err != nil {
		t.Fatalf("CountBlocks: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (upsert should not create duplicate)", count)
	}
}

func TestUpsertRawBlock(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	raws := loadJson(t, "blocks.json")
	raw1 := raws["0x42957"]

	if err := dbc.UpsertRawBlock(ctx, raw1); err != nil {
		t.Fatalf("first UpsertRawBlock: %v", err)
	}

	count, _ := dbc.CountBlocks(ctx)
	if count != 1 {
		t.Errorf("count after first upsert = %d, want 1", count)
	}

	raw2 := raws["0x42957"]
	time.Sleep(10 * time.Millisecond)
	if err := dbc.UpsertRawBlock(ctx, raw2); err != nil {
		t.Fatalf("second UpsertRawBlock: %v", err)
	}

	count, _ = dbc.CountBlocks(ctx)
	if count != 1 {
		t.Errorf("count after second upsert = %d, want 1 (upsert should not create duplicate)", count)
	}
}

func TestDeleteBlock(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	var rpcBlock rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &rpcBlock); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block, err := NewBlockFromRPC(&rpcBlock)
	if err != nil {
		t.Fatalf("NewBlockFromRPC: %v", err)
	}

	if err := dbc.InsertBlock(ctx, block); err != nil {
		t.Fatalf("InsertBlock: %v", err)
	}

	countBefore, _ := dbc.CountBlocks(ctx)
	if countBefore != 1 {
		t.Errorf("count before delete = %d, want 1", countBefore)
	}

	if err := dbc.DeleteBlock(ctx, 0x42957); err != nil {
		t.Fatalf("DeleteBlock: %v", err)
	}

	fetched, err := dbc.GetBlock(ctx, 0x42957)
	if err != nil {
		t.Fatalf("GetBlock after delete: %v", err)
	}
	if fetched != nil {
		t.Error("expected nil block after delete")
	}

	countAfter, _ := dbc.CountBlocks(ctx)
	if countAfter != 0 {
		t.Errorf("count after delete = %d, want 0", countAfter)
	}
}

func TestDeleteBlock_NonExistent(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	err := dbc.DeleteBlock(ctx, 999999)
	if err != nil {
		t.Errorf("unexpected error deleting non-existent block: %v", err)
	}
}

func TestCountBlocks(t *testing.T) {
	dbc := newTestDbContext(t)
	ctx := context.Background()

	count, err := dbc.CountBlocks(ctx)
	if err != nil {
		t.Fatalf("CountBlocks: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	var raw1 rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x42957"], &raw1); err != nil {
		t.Fatalf("unmarshal 0x42957: %v", err)
	}
	block1, _ := NewBlockFromRPC(&raw1)
	if err := dbc.InsertBlock(ctx, block1); err != nil {
		t.Fatalf("InsertBlock 1: %v", err)
	}

	var raw2 rpc.Block
	if err := json.Unmarshal(loadJson(t, "blocks.json")["0x1a33b7"], &raw2); err != nil {
		t.Fatalf("unmarshal 0x1a33b7: %v", err)
	}
	block2, _ := NewBlockFromRPC(&raw2)
	if err := dbc.InsertBlock(ctx, block2); err != nil {
		t.Fatalf("InsertBlock 2: %v", err)
	}

	count, err = dbc.CountBlocks(ctx)
	if err != nil {
		t.Fatalf("CountBlocks after insert: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func newTestDbContext(t *testing.T) *DbContext {
	t.Helper()
	dbc := &DbContext{
		uri:    dbUri,
		dbName: dbName,
	}
	if err := dbc.Connect(); err != nil {
		t.Fatalf("connect to mongodb: %v", err)
	}

	coll := dbc.db.Collection(BlockCollection)
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

