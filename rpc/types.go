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

const (
	TraceTypeDefault  = "default"
	TraceTypeCall     = "call"
	TraceTypePrestate = "prestate"
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
	LogsBloom        Bytes           `json:"logsBloom"`
	ParentHash       Bytes32         `json:"parentHash"`
	UncleHash        Bytes32         `json:"sha3Uncles"`
	TransactionsRoot Bytes32         `json:"transactionsRoot"`
	ReceiptsRoot     Bytes32         `json:"receiptsRoot"`
	StateRoot        Bytes32         `json:"stateRoot"`
	RawTransactions  json.RawMessage `json:"transactions,omitempty"`
	RawUncles        json.RawMessage `json:"uncles,omitempty"`
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

// CallFrame represents a single call frame of call tracer.
type CallFrame struct {
	From    Address     `json:"from"`
	Gas     Integer     `json:"gas"`
	GasUsed Integer     `json:"gasUsed"`
	To      Address     `json:"to"`
	Input   Bytes       `json:"input"`
	Output  Bytes       `json:"output,omitempty"`
	Value   Integer     `json:"value"`
	Type    string      `json:"type"`
	Calls   []CallFrame `json:"calls,omitempty"`
}

// StructLog represents a single EVM opcode execution step.
type StructLog struct {
	Pc      uint64            `json:"pc"`
	Op      string            `json:"op"`
	Gas     uint64            `json:"gas"`
	GasCost uint64            `json:"gasCost"`
	Depth   int               `json:"depth"`
	Stack   []string          `json:"stack"`
	Memory  []string          `json:"memory"`
	Storage map[string]string `json:"storage,omitempty"`
}

// PrestateAccount represents the prestate of an account for prestate rracer.
type PrestateAccount struct {
	Balance  Integer           `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Code     Bytes             `json:"code,omitempty"`
	CodeHash string            `json:"codeHash,omitempty"`
	Storage  map[string]string `json:"storage,omitempty"`
}

// Transaction represents an Ethereum transaction.
type Transaction struct {
	Hash        Bytes32 `json:"hash"`
	BlockNumber Integer `json:"blockNumber"`
	BlockHash   Bytes32 `json:"blockHash"`
	BlockTime   Integer `json:"blockTimestamp"`
	Nonce       Integer `json:"nonce"`
	Index       Integer `json:"transactionIndex"`
	From        Address `json:"from"`
	To          Address `json:"to"`
	Value       Integer `json:"value"`
	Gas         Integer `json:"gas"`
	GasPrice    Integer `json:"gasPrice"`
	Type        Integer `json:"type"`
	Input       Bytes   `json:"input"`
	V           Integer `json:"v"`
	R           Bytes   `json:"r"`
	S           Bytes   `json:"s"`
}

// TransactionTrace represents a single transaction trace.
type TransactionTrace struct {
	TxHash Bytes32         `json:"txHash"`
	Type   string          `json:"-"`
	Result json.RawMessage `json:"result"`
}

// Return call tracer result.
func (t *TransactionTrace) CallTraceResult() (*CallFrame, error) {
	if t.Type != TraceTypeCall {
		return nil, fmt.Errorf("trace type is %q, not %q", t.Type, TraceTypeCall)
	}
	var c CallFrame
	if err := json.Unmarshal(t.Result, &c); err != nil {
		return nil, fmt.Errorf("unmarshal call frame: %w", err)
	}
	return &c, nil
}

// Return default tracer result.
func (t *TransactionTrace) DefaultTraceResult() (*TransactionTraceResult, error) {
	if t.Type != TraceTypeDefault {
		return nil, fmt.Errorf("trace type is %q, not %q", t.Type, TraceTypeDefault)
	}
	var r TransactionTraceResult
	if err := json.Unmarshal(t.Result, &r); err != nil {
		return nil, fmt.Errorf("unmarshal structLogs result: %w", err)
	}
	return &r, nil
}

// Return prestate tracer result.
func (t *TransactionTrace) PrestateTraceResult() (map[string]PrestateAccount, error) {
	if t.Type != TraceTypePrestate {
		return nil, fmt.Errorf("trace type is %q, not %q", t.Type, TraceTypePrestate)
	}
	var m map[string]PrestateAccount
	if err := json.Unmarshal(t.Result, &m); err != nil {
		return nil, fmt.Errorf("unmarshal prestate: %w", err)
	}
	return m, nil
}

// TransactionTraceResult stores result of default tracer.
type TransactionTraceResult struct {
	Logs        []StructLog `json:"structLogs"`
	Gas         uint64      `json:"gas"`
	Failed      bool        `json:"failed"`
	ReturnValue string      `json:"returnValue"`
}
