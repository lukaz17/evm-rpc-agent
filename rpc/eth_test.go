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
	"testing"
)

const rpcEndpoint = "http://eth.tforce.xyz"

func newTestClient(t *testing.T) *Client {
	t.Helper()
	return NewClient(rpcEndpoint)
}

func TestBlockNumber(t *testing.T) {
	c := newTestClient(t)
	num, err := BlockNumber(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num == 0 {
		t.Fatal("expected block number > 0")
	}
	t.Logf("latest block number: %d", num)
}

func TestGetBlockByNumber_Earliest(t *testing.T) {
	c := newTestClient(t)
	block, err := GetBlockByNumber(c, "earliest", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block == nil {
		t.Fatal("expected non-nil block")
	}
	if len(block.Hash) == 0 {
		t.Fatal("expected non-empty earliest block hash")
	}
	genesis, _ := GetBlockByNumber(c, "0x0", false)
	if genesis != nil && block.Hash != genesis.Hash {
		t.Fatalf("earliest should match genesis: %s != %s", block.Hash, genesis.Hash)
	}
}

func TestGetBlockByNumber_FullTransactions(t *testing.T) {
	c := newTestClient(t)
	block, err := GetBlockByNumber(c, "latest", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block == nil {
		t.Fatal("expected non-nil block")
	}

	txs, err := block.TransactionsFull()
	if err != nil {
		t.Fatalf("unexpected error parsing full transactions: %v", err)
	}
	t.Logf("full transaction count: %d", len(txs))

	for i, tx := range txs {
		if len(tx.Hash) == 0 {
			t.Fatalf("tx[%d] has empty hash", i)
		}
		if len(tx.From) == 0 {
			t.Fatalf("tx[%d] has empty from address", i)
		}
	}
}

func TestGetBlockByNumber_Genesis(t *testing.T) {
	c := newTestClient(t)
	block, err := GetBlockByNumber(c, "0x0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block == nil {
		t.Fatal("expected non-nil genesis block")
	}
	if len(block.Hash) == 0 {
		t.Fatal("expected non-empty genesis block hash")
	}
	t.Logf("genesis block hash: %s", block.Hash)
}

func TestGetBlockByNumber_Latest(t *testing.T) {
	c := newTestClient(t)
	block, err := GetBlockByNumber(c, "latest", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block == nil {
		t.Fatal("expected non-nil block")
	}
	if len(block.Hash) == 0 {
		t.Fatal("expected non-empty block hash")
	}
	if block.Number.Hex() == "0x0" {
		t.Fatal("expected non-zero block number")
	}
	hashes, err := block.TransactionsHashes()
	if err != nil {
		t.Fatalf("unexpected error parsing transaction hashes: %v", err)
	}
	t.Logf("block %s, gasUsed=%d, txCount=%d", block.Number.Hex(), block.GasUsed.Uint64(), len(hashes))
}

func TestGetBlockByNumber_NonExistent(t *testing.T) {
	c := newTestClient(t)
	block, err := GetBlockByNumber(c, "0xffffffff", false)
	if block != nil {
		t.Fatal("expected nil block for non-existent number")
	}
	if err == nil {
		t.Log("node returned nil without error (acceptable)")
		return
	}
	t.Logf("got expected error: %v", err)
}

func TestGetBlockByNumber_Pending(t *testing.T) {
	c := newTestClient(t)
	block, err := GetBlockByNumber(c, "pending", false)
	if err != nil {
		t.Logf("node does not support pending: %v", err)
		return
	}
	if block == nil {
		t.Log("no pending blocks (expected on some nodes)")
		return
	}
	t.Logf("pending block hash: %s", block.Hash)
}
