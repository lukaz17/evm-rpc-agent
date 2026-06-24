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

	"github.com/lukaz17/evm-rpc-agent/core"
)

const (
	TraceTypeDefault  = "default"
	TraceTypeCall     = "call"
	TraceTypePrestate = "prestate"
)

// Block represents an Ethereum block.
type Block struct {
	Number           core.Integer    `json:"number"`
	Time             core.Integer    `json:"timestamp"`
	Hash             core.Bytes32    `json:"hash"`
	Nonce            core.Integer    `json:"nonce"`
	Coinbase         core.Address    `json:"miner"`
	Difficulty       core.Integer    `json:"difficulty"`
	TotalDifficulty  core.Integer    `json:"totalDifficulty"`
	Size             core.Integer    `json:"size"`
	GasUsed          core.Integer    `json:"gasUsed"`
	GasLimit         core.Integer    `json:"gasLimit"`
	Extra            core.Bytes      `json:"extraData"`
	MixDigest        core.Bytes32    `json:"mixHash"`
	LogsBloom        core.Bytes      `json:"logsBloom"`
	ParentHash       core.Bytes32    `json:"parentHash"`
	UncleHash        core.Bytes32    `json:"sha3Uncles"`
	TransactionsRoot core.Bytes32    `json:"transactionsRoot"`
	ReceiptsRoot     core.Bytes32    `json:"receiptsRoot"`
	StateRoot        core.Bytes32    `json:"stateRoot"`
	RawTransactions  json.RawMessage `json:"transactions,omitempty"`
	RawUncles        json.RawMessage `json:"uncles,omitempty"`
}

// Return transaction hashes from the block (when txDetails=false).
func (b *Block) TransactionsHashes() ([]core.Bytes32, error) {
	if b.RawTransactions == nil {
		return nil, nil
	}
	var hashes []core.Bytes32
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
func (b *Block) Uncles() ([]core.Bytes32, error) {
	if b.RawUncles == nil {
		return nil, nil
	}
	var uncles []core.Bytes32
	if err := json.Unmarshal(b.RawUncles, &uncles); err != nil {
		return nil, fmt.Errorf("unmarshal uncles: %w", err)
	}
	return uncles, nil
}

// CallFrame represents a single call frame of call tracer.
type CallFrame struct {
	From    core.Address  `json:"from"`
	Gas     core.Integer  `json:"gas"`
	GasUsed core.Integer  `json:"gasUsed"`
	To      core.Address  `json:"to"`
	Input   core.Bytes    `json:"input"`
	Output  core.Bytes    `json:"output,omitempty"`
	Value   core.Integer  `json:"value"`
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
	Balance  core.Integer      `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Code     core.Bytes        `json:"code,omitempty"`
	CodeHash string            `json:"codeHash,omitempty"`
	Storage  map[string]string `json:"storage,omitempty"`
}

// Transaction represents an Ethereum transaction.
type Transaction struct {
	Hash        core.Bytes32 `json:"hash"`
	BlockNumber core.Integer `json:"blockNumber"`
	BlockHash   core.Bytes32 `json:"blockHash"`
	BlockTime   core.Integer `json:"blockTimestamp"`
	Nonce       core.Integer `json:"nonce"`
	Index       core.Integer `json:"transactionIndex"`
	From        core.Address `json:"from"`
	To          core.Address `json:"to"`
	Value       core.Integer `json:"value"`
	Gas         core.Integer `json:"gas"`
	GasPrice    core.Integer `json:"gasPrice"`
	Type        core.Integer `json:"type"`
	Input       core.Bytes   `json:"input"`
	V           core.Integer `json:"v"`
	R           core.Bytes   `json:"r"`
	S           core.Bytes   `json:"s"`
}

// TransactionTrace represents a single transaction trace.
type TransactionTrace struct {
	TxHash core.Bytes32    `json:"txHash"`
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
