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

package rpc

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

func loadBlocksJSON(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	rel := filepath.Join("blobs", "blocks.json")
	data, err := os.ReadFile(rel)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal blocks.json: %v", err)
	}
	return raw
}

func TestBlock_UnmarshalJSON(t *testing.T) {
	rawBlocks := loadBlocksJSON(t)

	tests := []struct {
		name         string
		key          string
		wantHash     string
		wantNumber   uint64
		wantGasUsed  uint64
		wantTxCount  int
	}{
		{
			name:        "empty tx block",
			key:         "0x42957",
			wantHash:    "0x58e72ae0b02c13575f9124ed304cea92c1cfc8c9f8ed905d3bb2b57028a1e123",
			wantNumber:  0x42957,
			wantGasUsed: 0,
			wantTxCount: 0,
		},
		{
			name:        "block with transactions",
			key:         "0x1a33b7",
			wantHash:    "0x463b405b6b89683dd05956aaed190b3ab63c2d0efdbf1a99c5cd9f5aaeab11fc",
			wantNumber:  0x1a33b7,
			wantGasUsed: 0x63003,
			wantTxCount: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var block Block
			if err := json.Unmarshal(rawBlocks[tt.key], &block); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if block.Hash.Hex() != tt.wantHash {
				t.Errorf("hash = %s, want %s", block.Hash.Hex(), tt.wantHash)
			}
			if block.Number.Uint64() != tt.wantNumber {
				t.Errorf("number = 0x%x, want 0x%x", block.Number.Uint64(), tt.wantNumber)
			}
			if block.GasUsed.Uint64() != tt.wantGasUsed {
				t.Errorf("gasUsed = 0x%x, want 0x%x", block.GasUsed.Uint64(), tt.wantGasUsed)
			}

			txs, err := block.TransactionsFull()
			if err != nil {
				t.Fatalf("TransactionsFull: %v", err)
			}
			if len(txs) != tt.wantTxCount {
				t.Errorf("tx count = %d, want %d", len(txs), tt.wantTxCount)
			}
		})
	}
}

func TestBlock_FieldValues_EmptyTX(t *testing.T) {
	rawBlocks := loadBlocksJSON(t)

	var block Block
	if err := json.Unmarshal(rawBlocks["0x42957"], &block); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if block.GasLimit.Uint64() != 0x2fefd8 {
		t.Errorf("gasLimit = 0x%x, want 0x2fefd8", block.GasLimit.Uint64())
	}
	if block.Difficulty.Int().Cmp(big.NewInt(0x5fca12cbe53)) != 0 {
		t.Errorf("difficulty mismatch: %s", block.Difficulty.Hex())
	}
	if block.Coinbase.Hex() != "0xf8b483dba2c3b7176a3da549ad41a48bb3121069" {
		t.Errorf("coinbase = %s, want 0xf8b483dba2c3b7176a3da549ad41a48bb3121069", block.Coinbase.Hex())
	}
	if block.Extra.Hex() != "0x" {
		t.Errorf("extra = %s, want 0x", block.Extra.Hex())
	}
	if block.ParentHash.Hex() != "0xa8efedcde4c7a4267b3d465af3c4277d5e2735ada93f55ef7cca8b5d6a33f8d3" {
		t.Errorf("parentHash mismatch: %s", block.ParentHash.Hex())
	}
	if block.StateRoot.Hex() != "0x6081938266cf74834a3e36eff5fae4784fad8342fb283da53ed2718f46169d1f" {
		t.Errorf("stateRoot mismatch: %s", block.StateRoot.Hex())
	}

	hashes, err := block.TransactionsHashes()
	if err != nil {
		t.Fatalf("TransactionsHashes: %v", err)
	}
	if len(hashes) != 0 {
		t.Errorf("expected empty tx hashes, got %d", len(hashes))
	}

	uncles, err := block.Uncles()
	if err != nil {
		t.Fatalf("Uncles: %v", err)
	}
	if len(uncles) != 0 {
		t.Errorf("expected empty uncles, got %d", len(uncles))
	}
}

func TestBlock_FieldValues_WithTX(t *testing.T) {
	rawBlocks := loadBlocksJSON(t)

	var block Block
	if err := json.Unmarshal(rawBlocks["0x1a33b7"], &block); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if block.GasLimit.Uint64() != 0x47e7c4 {
		t.Errorf("gasLimit = 0x%x, want 0x47e7c4", block.GasLimit.Uint64())
	}
	if block.Coinbase.Hex() != "0x2a65aca4d5fc5b5c859090a6c34d164135398226" {
		t.Errorf("coinbase mismatch: %s", block.Coinbase.Hex())
	}
	if block.Extra.Hex() != "0xd783010406844765746887676f312e362e31856c696e7578" {
		t.Errorf("extra mismatch: %s", block.Extra.Hex())
	}

	txs, err := block.TransactionsFull()
	if err != nil {
		t.Fatalf("TransactionsFull: %v", err)
	}
	if len(txs) != 9 {
		t.Fatalf("tx count = %d, want 9", len(txs))
	}
}

func TestBlock_TransactionsHashes_EmptyArray(t *testing.T) {
	rawBlocks := loadBlocksJSON(t)

	var block Block
	if err := json.Unmarshal(rawBlocks["0x42957"], &block); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	hashes, err := block.TransactionsHashes()
	if err != nil {
		t.Fatalf("TransactionsHashes: %v", err)
	}
	if len(hashes) != 0 {
		t.Errorf("expected empty tx hashes, got %d", len(hashes))
	}
}

func TestBlock_TransactionsFull(t *testing.T) {
	rawBlocks := loadBlocksJSON(t)

	var block Block
	if err := json.Unmarshal(rawBlocks["0x1a33b7"], &block); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	txs, err := block.TransactionsFull()
	if err != nil {
		t.Fatalf("TransactionsFull: %v", err)
	}

	if len(txs) != 9 {
		t.Fatalf("tx count = %d, want 9", len(txs))
	}

	first := txs[0]
	if first.Hash.Hex() != "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686" {
		t.Errorf("first tx hash mismatch: %s", first.Hash.Hex())
	}
	if first.From.Hex() != "0xfbe26da0e985087d28228defbdaa394713b0865f" {
		t.Errorf("first tx from mismatch: %s", first.From.Hex())
	}
	if first.To == nil || first.To.Hex() != "0xfb110ca742c3e47ab3babf4d48624c33891f35c6" {
		t.Errorf("first tx to mismatch: %v", first.To)
	}
	if first.Value.Int().Cmp(big.NewInt(0x1062fab684fee0)) != 0 {
		t.Errorf("first tx value = %s, want 0x1062fab684fee0", first.Value.Hex())
	}
	if first.GasPrice.Int().Cmp(big.NewInt(0x9502f9000)) != 0 {
		t.Errorf("first tx gasPrice = %s, want 0x9502f9000", first.GasPrice.Hex())
	}
	if first.R.Hex() != "0x4b089b013cad39edd3fb1ee8dbe36e131a570aa807baea80efb8c61b2967d396" {
		t.Errorf("first tx r mismatch: %s", first.R.Hex())
	}
	if first.S.Hex() != "0x6d2ef2f6135f38c61ca13b340265fc587c2dcec980c24294807b105279b4efca" {
		t.Errorf("first tx s mismatch: %s", first.S.Hex())
	}

	last := txs[8]
	if last.Hash.Hex() != "0xb1bcf68cec45d5f3dde9e21b8b4afd3131f933373bb85068bfc572dc79fc867b" {
		t.Errorf("last tx hash mismatch: %s", last.Hash.Hex())
	}
	if last.Input.Hex() != "0x667a2f58" {
		t.Errorf("last tx input = %s, want 0x667a2f58", last.Input.Hex())
	}
	if last.Value.Int().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("last tx value = %s, want 0x0", last.Value.Hex())
	}
}

func TestTransaction_UnmarshalJSON(t *testing.T) {
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

	var tx Transaction
	if err := json.Unmarshal([]byte(txJSON), &tx); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if tx.Hash.Hex() != "0xc2c239e7c20fd3b5dccfc42e91bd4ae4a68727cfe184af3ea1ad7824cc13a686" {
		t.Errorf("hash mismatch: %s", tx.Hash.Hex())
	}
	if tx.BlockNumber == nil || tx.BlockNumber.Uint64() != 0x1a33b7 {
		t.Errorf("blockNumber = %v, want 0x1a33b7", tx.BlockNumber)
	}
	if tx.BlockHash == nil || tx.BlockHash.Hex() != "0x463b405b6b89683dd05956aaed190b3ab63c2d0efdbf1a99c5cd9f5aaeab11fc" {
		t.Errorf("blockHash mismatch: %v", tx.BlockHash)
	}
	if tx.Index == nil || tx.Index.Uint64() != 0 {
		t.Errorf("index = %v, want 0", tx.Index)
	}
	if tx.From.Hex() != "0xfbe26da0e985087d28228defbdaa394713b0865f" {
		t.Errorf("from mismatch: %s", tx.From.Hex())
	}
	if tx.To == nil || tx.To.Hex() != "0xfb110ca742c3e47ab3babf4d48624c33891f35c6" {
		t.Errorf("to mismatch: %v", tx.To)
	}
	if tx.Gas.Uint64() != 0x5208 {
		t.Errorf("gas = 0x%x, want 0x5208", tx.Gas.Uint64())
	}
	if tx.Nonce.Int().Cmp(big.NewInt(0x1390)) != 0 {
		t.Errorf("nonce = %s, want 0x1390", tx.Nonce.Hex())
	}
	if tx.Type.Int().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("type = %s, want 0x0", tx.Type.Hex())
	}
	if tx.Input.Hex() != "0x" {
		t.Errorf("input = %s, want 0x", tx.Input.Hex())
	}
}

func TestTransaction_UnmarshalJSON_WithInput(t *testing.T) {
	txJSON := `{
		"blockHash": "0x463b405b6b89683dd05956aaed190b3ab63c2d0efdbf1a99c5cd9f5aaeab11fc",
		"blockNumber": "0x1a33b7",
		"blockTimestamp": "0x57632695",
		"from": "0xed059bc543141c8c93031d545079b3da0233b27f",
		"gas": "0x3d090",
		"gasPrice": "0x4a817c800",
		"hash": "0xb1bcf68cec45d5f3dde9e21b8b4afd3131f933373bb85068bfc572dc79fc867b",
		"input": "0x667a2f58",
		"nonce": "0x34b1",
		"to": "0x8b3b3b624c3c0397d3da8fd861512393d51dcbac",
		"transactionIndex": "0x8",
		"value": "0x0",
		"type": "0x0",
		"v": "0x1b",
		"r": "0xcd858ea2aeb7da2a77f970aa382994f6b4de55108d2ff329e3973c0f94808348",
		"s": "0x5d567aec7befb564b844e81959a6a27f97a4c248cd7a17755230b00b3efb9167"
	}`

	var tx Transaction
	if err := json.Unmarshal([]byte(txJSON), &tx); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if tx.Input.Hex() != "0x667a2f58" {
		t.Errorf("input = %s, want 0x667a2f58", tx.Input.Hex())
	}
	if tx.Gas.Uint64() != 0x3d090 {
		t.Errorf("gas = 0x%x, want 0x3d090", tx.Gas.Uint64())
	}
	if tx.Index == nil || tx.Index.Uint64() != 8 {
		t.Errorf("index = %v, want 8", tx.Index)
	}
	if tx.V.Int().Cmp(big.NewInt(0x1b)) != 0 {
		t.Errorf("v = %s, want 0x1b", tx.V.Hex())
	}
}

func TestBlock_RoundTrip(t *testing.T) {
	rawBlocks := loadBlocksJSON(t)

	for key, raw := range rawBlocks {
		t.Run(key, func(t *testing.T) {
			var block1 Block
			if err := json.Unmarshal(raw, &block1); err != nil {
				t.Fatalf("first unmarshal: %v", err)
			}

			marshaled, err := json.Marshal(&block1)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var block2 Block
			if err := json.Unmarshal(marshaled, &block2); err != nil {
				t.Fatalf("second unmarshal: %v", err)
			}

			if block1.Hash.Hex() != block2.Hash.Hex() {
				t.Errorf("hash mismatch after round-trip")
			}
			if block1.Number.Uint64() != block2.Number.Uint64() {
				t.Errorf("number mismatch after round-trip")
			}
			if block1.GasUsed.Int().Cmp(block2.GasUsed.Int()) != 0 {
				t.Errorf("gasUsed mismatch after round-trip")
			}
			if block1.Difficulty.Int().Cmp(block2.Difficulty.Int()) != 0 {
				t.Errorf("difficulty mismatch after round-trip")
			}

			txs1, err := block1.TransactionsFull()
			if err != nil {
				t.Fatalf("block1 TransactionsFull: %v", err)
			}
			txs2, err := block2.TransactionsFull()
			if err != nil {
				t.Fatalf("block2 TransactionsFull: %v", err)
			}
			if len(txs1) != len(txs2) {
				t.Errorf("tx count mismatch after round-trip: %d vs %d", len(txs1), len(txs2))
			}
			for i := range txs1 {
				if txs1[i].Hash.Hex() != txs2[i].Hash.Hex() {
					t.Errorf("tx[%d] hash mismatch after round-trip", i)
				}
				if txs1[i].From.Hex() != txs2[i].From.Hex() {
					t.Errorf("tx[%d] from mismatch after round-trip", i)
				}
			}
		})
	}
}

func TestTransaction_RoundTrip(t *testing.T) {
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

	var tx1 Transaction
	if err := json.Unmarshal([]byte(txJSON), &tx1); err != nil {
		t.Fatalf("first unmarshal: %v", err)
	}

	marshaled, err := json.Marshal(&tx1)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var tx2 Transaction
	if err := json.Unmarshal(marshaled, &tx2); err != nil {
		t.Fatalf("second unmarshal: %v", err)
	}

	if tx1.Hash.Hex() != tx2.Hash.Hex() {
		t.Errorf("hash mismatch after round-trip")
	}
	if tx1.From.Hex() != tx2.From.Hex() {
		t.Errorf("from mismatch after round-trip")
	}
	if (tx1.To == nil) != (tx2.To == nil) || (tx1.To != nil && tx1.To.Hex() != tx2.To.Hex()) {
		t.Errorf("to mismatch after round-trip")
	}
	if tx1.Value.Int().Cmp(tx2.Value.Int()) != 0 {
		t.Errorf("value mismatch after round-trip")
	}
	if tx1.GasPrice.Int().Cmp(tx2.GasPrice.Int()) != 0 {
		t.Errorf("gasPrice mismatch after round-trip")
	}
	if tx1.R.Hex() != tx2.R.Hex() {
		t.Errorf("r mismatch after round-trip")
	}
	if tx1.S.Hex() != tx2.S.Hex() {
		t.Errorf("s mismatch after round-trip")
	}
}
