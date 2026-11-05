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
	if block.Data.Hash.Hex() != "0x58e72ae0b02c13575f9124ed304cea92c1cfc8c9f8ed905d3bb2b57028a1e123" {
		t.Errorf("Hash = %s", block.Data.Hash)
	}
	if block.Data.Miner.Hex() != "0xf8b483dba2c3b7176a3da549ad41a48bb3121069" {
		t.Errorf("Miner = %s", block.Data.Miner)
	}
	if block.Data.GasUsed.Uint64() != 0 {
		t.Errorf("GasUsed = %d, want 0", block.Data.GasUsed.Uint64())
	}
	if block.Data.GasLimit.Uint64() != 0x2fefd8 {
		t.Errorf("GasLimit = 0x%x, want 0x2fefd8", block.Data.GasLimit.Uint64())
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
	if block.Data.Hash.Hex() != "0x463b405b6b89683dd05956aaed190b3ab63c2d0efdbf1a99c5cd9f5aaeab11fc" {
		t.Errorf("Hash = %s", block.Data.Hash)
	}
	if block.Data.GasUsed.Uint64() != 0x63003 {
		t.Errorf("GasUsed = 0x%x, want 0x63003", block.Data.GasUsed.Uint64())
	}

	if len(block.Data.Transactions) != 9 {
		t.Fatalf("Transactions count = %d, want 9", len(block.Data.Transactions))
	}

	firstTx := block.Data.Transactions[0]
	if firstTx.Hash.Hex() != "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686" {
		t.Errorf("first tx hash = %s", firstTx.Hash)
	}
	if firstTx.From.Hex() != "0xfbe26da0e985087d28228defbdaa394713b0865f" {
		t.Errorf("first tx from = %s", firstTx.From)
	}
	if firstTx.To.Hex() != "0xfb110ca742c3e47ab3babf4d48624c33891f35c6" {
		t.Errorf("first tx to = %v", firstTx.To)
	}
	if firstTx.Gas.Uint64() != 0x5208 {
		t.Errorf("first tx gas = 0x%x, want 0x5208", firstTx.Gas.Uint64())
	}

	lastTx := block.Data.Transactions[8]
	if lastTx.Input.Hex() != "0x667a2f58" {
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

	if tx.Hash.Hex() != "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686" {
		t.Errorf("hash = %s", tx.Hash)
	}
	if tx.BlockNumber.Uint64() != 0x1a33b7 {
		t.Errorf("blockNumber = %d, want 0x1a33b7", tx.BlockNumber.Uint64())
	}
	if tx.From.Hex() != "0xfbe26da0e985087d28228defbdaa394713b0865f" {
		t.Errorf("from = %s", tx.From)
	}
	if tx.To.Hex() != "0xfb110ca742c3e47ab3babf4d48624c33891f35c6" {
		t.Errorf("to = %v", tx.To)
	}
	if tx.Gas.Uint64() != 0x5208 {
		t.Errorf("gas = 0x%x, want 0x5208", tx.Gas.Uint64())
	}
	if tx.Index.Uint64() != 0 {
		t.Errorf("index = %d, want 0", tx.Index.Uint64())
	}
}
