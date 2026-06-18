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
	"fmt"
)

// Block represents an Ethereum block.
type Block struct {
	Number           Integer         `json:"number"`
	Time             Integer         `json:"timestamp"`
	Hash             Bytes32         `json:"hash"`
	Nonce            Integer         `json:"nonce"`
	Coinbase         Address         `json:"miner"`
	Difficulty       Integer         `json:"difficulty"`
	TotalDifficulty  Integer         `json:"totalDifficulty"`
	Size             Integer         `json:"size"`
	GasUsed          Integer         `json:"gasUsed"`
	GasLimit         Integer         `json:"gasLimit"`
	Extra            Bytes           `json:"extraData"`
	MixDigest        Bytes32         `json:"mixHash"`
	Bloom            Bytes           `json:"logsBloom"`
	ParentHash       Bytes32         `json:"parentHash"`
	UncleHash        Bytes32         `json:"sha3Uncles"`
	TransactionsRoot Bytes32         `json:"transactionsRoot"`
	ReceiptsRoot     Bytes32         `json:"receiptsRoot"`
	StateRoot        Bytes32         `json:"stateRoot"`
	RawTransactions  json.RawMessage `json:"transactions"`
	RawUncles        json.RawMessage `json:"uncles"`
}

// Return transaction hashes from the block (when txDetails=false).
func (b *Block) TransactionsHashes() ([]Bytes32, error) {
	if b.RawTransactions == nil {
		return nil, nil
	}
	var hashes []Bytes32
	if err := json.Unmarshal(b.RawTransactions, &hashes); err != nil {
		return nil, fmt.Errorf("unmarshal transaction hashes: %w", err)
	}
	return hashes, nil
}

// Return full transaction objects from the block (when txDetails=true).
func (b *Block) TransactionsFull() ([]Transaction, error) {
	if b.RawTransactions == nil {
		return nil, nil
	}
	var txs []Transaction
	if err := json.Unmarshal(b.RawTransactions, &txs); err != nil {
		return nil, fmt.Errorf("unmarshal transactions: %w", err)
	}
	return txs, nil
}

// Return uncle hashes from the block.
func (b *Block) Uncles() ([]Bytes32, error) {
	if b.RawUncles == nil {
		return nil, nil
	}
	var uncles []Bytes32
	if err := json.Unmarshal(b.RawUncles, &uncles); err != nil {
		return nil, fmt.Errorf("unmarshal uncles: %w", err)
	}
	return uncles, nil
}

// Transaction represents an Ethereum transaction.
type Transaction struct {
	Hash        Bytes32  `json:"hash"`
	BlockNumber *Integer `json:"blockNumber"`
	BlockHash   *Bytes32 `json:"blockHash"`
	BlockTime   Integer  `json:"blockTimestamp"`
	Nonce       Integer  `json:"nonce"`
	Index       *Integer `json:"transactionIndex"`
	From        Address  `json:"from"`
	To          *Address `json:"to"`
	Value       Integer  `json:"value"`
	Gas         Integer  `json:"gas"`
	GasPrice    Integer  `json:"gasPrice"`
	Type        Integer  `json:"type"`
	Input       Bytes    `json:"input"`
	V           Integer  `json:"v"`
	R           Bytes    `json:"r"`
	S           Bytes    `json:"s"`
}
