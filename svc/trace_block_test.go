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

package svc

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/lukaz17/evm-rpc-agent/mem"
	"github.com/tforce-io/tf-golib/multiplex"
)

func TestTraceBlock_DownloadTraces(t *testing.T) {
	fromBlock := big.NewInt(200)
	toBlock := big.NewInt(208)
	batchSize := 4
	expectedCount := int(toBlock.Int64()-fromBlock.Int64()) + 1

	dbc := newTestDbContext(t, []string{db.CallTraceCollection})
	ctrl := newTestController(t, dbc)

	params := multiplex.ExecParams{
		"from_block_number": fromBlock,
		"to_block_number":   toBlock,
		"batch_size":        batchSize,
	}
	ctrl.DispatchOnce("TraceBlock", "trace_calls", params)

	result := params.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}

	time.Sleep(500 * time.Millisecond)

	finalCount, err := dbc.CountCallTraces(context.Background())
	if err != nil {
		t.Fatalf("count call traces: %v", err)
	}

	if int(finalCount) < expectedCount {
		t.Errorf("expected at least %d new call traces in DB, got %d", expectedCount, finalCount)
	}

	for blockNum := fromBlock.Int64(); blockNum <= toBlock.Int64(); blockNum++ {
		trace, err := dbc.GetCallTrace(context.Background(), uint64(blockNum))
		if err != nil {
			t.Errorf("get call trace %d: %v", blockNum, err)
			continue
		}

		if trace == nil {
			t.Errorf("call trace %d not found in database", blockNum)
			continue
		}

		if trace.Number != uint64(blockNum) {
			t.Errorf("trace number = %d, want %d", trace.Number, blockNum)
		}
	}

	if mem.CurrentHeight.CallTrace != toBlock.Uint64() {
		t.Errorf("mem.CurrentHeight.CallTrace = %d, want %d", mem.CurrentHeight.CallTrace, toBlock.Uint64())
	}
}

func TestTraceBlock_BatchProcessing(t *testing.T) {
	fromBlock := big.NewInt(300)
	toBlock := big.NewInt(307)
	batchSize := 3
	expectedCount := int(toBlock.Int64()-fromBlock.Int64()) + 1

	dbc := newTestDbContext(t, []string{db.CallTraceCollection})
	ctrl := newTestController(t, dbc)

	params := multiplex.ExecParams{
		"from_block_number": fromBlock,
		"to_block_number":   toBlock,
		"batch_size":        batchSize,
	}
	ctrl.DispatchOnce("TraceBlock", "trace_calls", params)

	result := params.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}

	time.Sleep(500 * time.Millisecond)

	finalCount, err := dbc.CountCallTraces(context.Background())
	if err != nil {
		t.Fatalf("count call traces: %v", err)
	}

	if int(finalCount) < expectedCount {
		t.Errorf("expected at least %d new call traces in DB (4 batches of 3), got %d", expectedCount, finalCount)
	}

	for blockNum := fromBlock.Int64(); blockNum <= toBlock.Int64(); blockNum++ {
		trace, err := dbc.GetCallTrace(context.Background(), uint64(blockNum))
		if err != nil {
			t.Errorf("get call trace %d: %v", blockNum, err)
			continue
		}

		if trace == nil {
			t.Errorf("call trace %d not found in database", blockNum)
			continue
		}
	}
}

func TestTraceBlock_SingleBlock(t *testing.T) {
	block := big.NewInt(400)

	dbc := newTestDbContext(t, []string{db.CallTraceCollection})
	ctrl := newTestController(t, dbc)

	params := multiplex.ExecParams{
		"from_block_number": block,
		"to_block_number":   block,
		"batch_size":        1,
	}
	ctrl.DispatchOnce("TraceBlock", "trace_calls", params)

	result := params.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}

	time.Sleep(500 * time.Millisecond)

	trace, err := dbc.GetCallTrace(context.Background(), block.Uint64())
	if err != nil {
		t.Fatalf("get call trace %d: %v", block.Int64(), err)
	}

	if trace == nil {
		t.Fatalf("call trace %d not found in database", block.Int64())
	}

	if trace.Number != block.Uint64() {
		t.Errorf("trace number = %d, want %d", trace.Number, block.Uint64())
	}

	if mem.CurrentHeight.CallTrace != block.Uint64() {
		t.Errorf("mem.CurrentHeight.CallTrace = %d, want %d", mem.CurrentHeight.CallTrace, block.Uint64())
	}
}

func TestTraceBlock_ParamsExtraction(t *testing.T) {
	tests := []struct {
		name      string
		params    multiplex.ExecParams
		wantFrom  int64
		wantTo    int64
		wantBatch int
	}{
		{
			name:      "default values",
			params:    multiplex.ExecParams{},
			wantFrom:  0,
			wantTo:    0,
			wantBatch: 1,
		},
		{
			name: "explicit values",
			params: multiplex.ExecParams{
				"from_block_number": big.NewInt(1000),
				"to_block_number":   big.NewInt(2000),
				"batch_size":        50,
			},
			wantFrom:  1000,
			wantTo:    2000,
			wantBatch: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &multiplex.ServiceMessage{
				Params: tt.params,
			}
			p := NewTraceBlockParams(msg)

			if p.FromBlockNumber.Int64() != tt.wantFrom {
				t.Errorf("from_block_number = %d, want %d", p.FromBlockNumber.Int64(), tt.wantFrom)
			}
			if p.ToBlockNumber.Int64() != tt.wantTo {
				t.Errorf("to_block_number = %d, want %d", p.ToBlockNumber.Int64(), tt.wantTo)
			}
			if p.BatchSize != tt.wantBatch {
				t.Errorf("batch_size = %d, want %d", p.BatchSize, tt.wantBatch)
			}
		})
	}
}

func TestTraceBlock_MidBatchError(t *testing.T) {
	fromBlock := big.NewInt(500)
	toBlock := big.NewInt(507)
	batchSize := 4

	dbc := newTestDbContext(t, []string{db.CallTraceCollection})
	ctrl := newTestController(t, dbc)

	params := multiplex.ExecParams{
		"from_block_number": fromBlock,
		"to_block_number":   toBlock,
		"batch_size":        batchSize,
	}
	ctrl.DispatchOnce("TraceBlock", "trace_calls", params)

	result := params.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}

	time.Sleep(500 * time.Millisecond)

	finalCount, err := dbc.CountCallTraces(context.Background())
	if err != nil {
		t.Fatalf("count call traces: %v", err)
	}

	if finalCount == 0 {
		t.Fatal("expected at least some call traces to be written before any error")
	}

	trace, err := dbc.GetCallTrace(context.Background(), fromBlock.Uint64())
	if err != nil || trace == nil {
		t.Errorf("first block %d should have been traced", fromBlock.Int64())
	}

	if mem.CurrentHeight.CallTrace < fromBlock.Uint64() {
		t.Errorf("mem.CurrentHeight.CallTrace = %d, want >= %d", mem.CurrentHeight.CallTrace, fromBlock.Uint64())
	}
}
