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
	"testing"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

func TestCallEthRpc_EthGetBlockByNumber(t *testing.T) {
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

	msg := multiplex.ExecParams{
		"method": "eth_getBlockByNumber",
		"params": []any{"0x1a33b7", true},
	}
	ctrl.DispatchOnce("StdCallEthRpc", "rpc_call", msg)

	result, ok := msg.ReturnResult().(*CallEthRpcResult)
	if !ok {
		t.Fatalf("result is not *CallEthRpcResult, got %T", result)
	}

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if result.Method != "eth_getBlockByNumber" {
		t.Errorf("method = %q, want %q", result.Method, "eth_getBlockByNumber")
	}

	var block map[string]json.RawMessage
	if err := json.Unmarshal(result.Data, &block); err != nil {
		t.Fatalf("unmarshal block data: %v", err)
	}

	numRaw, ok := block["number"]
	if !ok {
		t.Fatal("missing 'number' field in block response")
	}

	var numHex string
	if err := json.Unmarshal(numRaw, &numHex); err != nil {
		t.Fatalf("unmarshal block number: %v", err)
	}

	if numHex != "0x1a33b7" {
		t.Errorf("block number = %q, want %q", numHex, "0x1a33b7")
	}
}
