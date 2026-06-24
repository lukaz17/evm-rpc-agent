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
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

func TestCallEthApi_GetBlocks(t *testing.T) {
	blockNumbers := []*big.Int{
		big.NewInt(17),
		big.NewInt(171),
		big.NewInt(1717),
		big.NewInt(2727),
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: &nullWriter{}, TimeFormat: time.DateTime}).
		With().
		Timestamp().
		Logger()
	rpcClient := rpc.NewClient(rpcEndpoint)

	cfg := &config.ServiceConfig{
		MaxRpcRrtryCount:         3,
		HistoricalApiWorkerCount: 1,
		StandardApiWorkerCount:   1,
	}
	ctrl := NewController(cfg, rpcClient, logger)
	go ctrl.Run()

	params := multiplex.ExecParams{
		"block_numbers": blockNumbers,
	}
	ctrl.DispatchOnce("CallEthApi", "get_blocks", params)

	result, ok := params.ReturnResult().(*CallEthApiResult)
	if !ok {
		t.Fatalf("result is not *CallEthApiResult, got %T", result)
	}

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Data) != len(blockNumbers) {
		t.Fatalf("data length = %d, want %d", len(result.Data), len(blockNumbers))
	}

	for i, item := range result.Data {
		expectedID := blockNumbers[i].Text(10)
		if item.ID != expectedID {
			t.Errorf("item[%d] ID = %q, want %q", i, item.ID, expectedID)
		}

		if item.Error != nil {
			t.Errorf("item[%d] unexpected error: %v", i, item.Error)
		}

		var block map[string]json.RawMessage
		if err := json.Unmarshal([]byte(item.Data), &block); err != nil {
			t.Fatalf("unmarshal block data for ID %s: %v", item.ID, err)
		}

		numRaw, ok := block["number"]
		if !ok {
			t.Fatalf("missing 'number' field in block response for ID %s", item.ID)
		}

		var numHex string
		if err := json.Unmarshal(numRaw, &numHex); err != nil {
			t.Fatalf("unmarshal block number for ID %s: %v", item.ID, err)
		}

		expectedHex := "0x" + blockNumbers[i].Text(16)
		if numHex != expectedHex {
			t.Errorf("block number = %q, want %q", numHex, expectedHex)
		}
	}
}

func TestCallEthApi_GetBlocksRange(t *testing.T) {
	fromBlock := big.NewInt(101)
	toBlock := big.NewInt(117)
	expectedCount := int(toBlock.Int64()-fromBlock.Int64()) + 1

	logger := zerolog.New(zerolog.ConsoleWriter{Out: &nullWriter{}, TimeFormat: time.DateTime}).
		With().
		Timestamp().
		Logger()
	rpcClient := rpc.NewClient(rpcEndpoint)

	cfg := &config.ServiceConfig{
		MaxRpcRrtryCount:         3,
		HistoricalApiWorkerCount: 1,
		StandardApiWorkerCount:   1,
	}
	ctrl := NewController(cfg, rpcClient, logger)
	go ctrl.Run()

	params := multiplex.ExecParams{
		"from_block_number": fromBlock,
		"to_block_number":   toBlock,
	}
	ctrl.DispatchOnce("CallEthApi", "get_blocks_range", params)

	result, ok := params.ReturnResult().(*CallEthApiResult)
	if !ok {
		t.Fatalf("result is not *CallEthApiResult, got %T", result)
	}

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Data) != expectedCount {
		t.Fatalf("data length = %d, want %d", len(result.Data), expectedCount)
	}

	for i, item := range result.Data {
		expectedBlockNum := fromBlock.Int64() + int64(i)
		expectedID := big.NewInt(expectedBlockNum).Text(10)
		if item.ID != expectedID {
			t.Errorf("item[%d] ID = %q, want %q", i, item.ID, expectedID)
		}

		if item.Error != nil {
			t.Errorf("item[%d] unexpected error: %v", i, item.Error)
		}

		var block map[string]json.RawMessage
		if err := json.Unmarshal([]byte(item.Data), &block); err != nil {
			t.Fatalf("unmarshal block data for ID %s: %v", item.ID, err)
		}

		numRaw, ok := block["number"]
		if !ok {
			t.Fatalf("missing 'number' field in block response for ID %s", item.ID)
		}

		var numHex string
		if err := json.Unmarshal(numRaw, &numHex); err != nil {
			t.Fatalf("unmarshal block number for ID %s: %v", item.ID, err)
		}

		expectedHex := "0x" + big.NewInt(expectedBlockNum).Text(16)
		if numHex != expectedHex {
			t.Errorf("block number = %q, want %q", numHex, expectedHex)
		}
	}
}
