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
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/core"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

// CallEthApi executes high-level Ethereum API with batch and paging support.
type CallEthApi struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
	o *config.ServiceConfig
}

// Return new CallEthApi instance.
func NewCallEthApi(cfg *config.ServiceConfig, logger zerolog.Logger) *CallEthApi {
	svc := &CallEthApi{
		o: cfg,
	}
	svc.i = svc.InitServiceCore("CallEthApi", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *CallEthApi) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "block_number":
		s.i.Logger.Infof("%s#%02d: %s started.", s.i.ServiceID, workerID, msg.Command)
		request := multiplex.ExecParams{
			"method": "eth_blockNumber",
			"params": []any{},
		}
		request.ExpectReturn()
		s.Dispatch("ExclCallEthRpc", "rpc_call", request)
		result := request.WaitForReturn().(*CallEthRpcResult)
		var data uint64
		if result.Error == nil {
			var hexStr string
			if err := json.Unmarshal(result.Data, &hexStr); err != nil {
				result.Error = fmt.Errorf("unmarshal block number: %w", err)
			} else {
				val, parseErr := core.DecodeNumericString(hexStr)
				if parseErr != nil {
					result.Error = fmt.Errorf("parse block number %s: %w", hexStr, parseErr)
				} else {
					data = val.Uint64()
				}
			}
		}
		msg.Return(&CallBlockNumberResult{
			Data:  data,
			Error: result.Error,
		})
	case "get_blocks":
		fallthrough
	case "get_blocks_range":
		s.i.Logger.Infof("%s#%02d: %s started.", s.i.ServiceID, workerID, msg.Command)
		startTime := time.Now()
		p := NewCallEthApiParams(msg)
		requests := []multiplex.ExecParams{}
		signal := new(sync.WaitGroup)
		if msg.Command == "get_blocks" {
			for _, blockNumber := range p.BlockNumbers {
				hexNum := "0x" + blockNumber.Text(16)
				request := multiplex.ExecParams{
					"method":       "eth_getBlockByNumber",
					"params":       []any{hexNum, true},
					"block_number": new(big.Int).Set(blockNumber),
					"signal":       signal,
				}
				request.ExpectReturnCustomSignal(signal)
				requests = append(requests, request)
			}
		}
		if msg.Command == "get_blocks_range" {
			for blockNumber := new(big.Int).Set(p.FromBlockNumber); blockNumber.Cmp(p.ToBlockNumber) <= 0; blockNumber.Set(new(big.Int).Add(blockNumber, big.NewInt(1))) {
				hexNum := "0x" + blockNumber.Text(16)
				request := multiplex.ExecParams{
					"method":       "eth_getBlockByNumber",
					"params":       []any{hexNum, true},
					"block_number": new(big.Int).Set(blockNumber),
				}
				request.ExpectReturnCustomSignal(signal)
				requests = append(requests, request)
			}
		}
		result := &CallEthApiResult{
			Data: make([]*CallEthApiItem, len(requests)),
		}
		errorCount := 0
		anomalyCount := 0
		if len(requests) > 0 {
			signal.Add(len(requests))
			for _, request := range requests {
				s.Dispatch("StdCallEthRpc", "rpc_call", request)
			}
			signal.Wait()
			for i, request := range requests {
				blockNumber := request["block_number"].(*big.Int)
				callRpcResult := request.ReturnResult().(*CallEthRpcResult)
				result.Data[i] = &CallEthApiItem{
					ID:    blockNumber.Text(10),
					Data:  string(callRpcResult.Data),
					Error: callRpcResult.Error,
				}
				if result.Data[i].Error != nil {
					errorCount++
					continue
				}
				anomalyCount += s.dispatchBlockAnomalies(workerID, blockNumber, callRpcResult.Data)
			}
			if errorCount > 0 {
				result.Error = errors.New("error while invoking api, check inner data for details")
			}
		}
		s.i.Logger.Infof("%s#%02d: %d blocks retrieved in %v. Error count = %d. Anomaly count = %d.", s.i.ServiceID, workerID, len(requests), time.Since(startTime), errorCount, anomalyCount)
		msg.Return(result)
	case "trace_call":
		fallthrough
	case "trace_calls_range":
		s.i.Logger.Infof("%s#%02d: %s started.", s.i.ServiceID, workerID, msg.Command)
		startTime := time.Now()
		p := NewCallEthApiParams(msg)
		requests := []multiplex.ExecParams{}
		signal := new(sync.WaitGroup)
		if msg.Command == "trace_call" {
			for _, blockNumber := range p.BlockNumbers {
				hexNum := "0x" + blockNumber.Text(16)
				tracerOption := traceBlockTracerOptions(s.o.TraceTimeoutSeconds)
				request := multiplex.ExecParams{
					"method":       "debug_traceBlockByNumber",
					"params":       []any{hexNum, tracerOption},
					"block_number": new(big.Int).Set(blockNumber),
					"signal":       signal,
				}
				request.ExpectReturnCustomSignal(signal)
				requests = append(requests, request)
			}
		}
		if msg.Command == "trace_calls_range" {
			for blockNumber := new(big.Int).Set(p.FromBlockNumber); blockNumber.Cmp(p.ToBlockNumber) <= 0; blockNumber.Set(new(big.Int).Add(blockNumber, big.NewInt(1))) {
				hexNum := "0x" + blockNumber.Text(16)
				tracerOption := traceBlockTracerOptions(s.o.TraceTimeoutSeconds)
				request := multiplex.ExecParams{
					"method":       "debug_traceBlockByNumber",
					"params":       []any{hexNum, tracerOption},
					"block_number": new(big.Int).Set(blockNumber),
				}
				request.ExpectReturnCustomSignal(signal)
				requests = append(requests, request)
			}
		}
		result := &CallEthApiResult{
			Data: make([]*CallEthApiItem, len(requests)),
		}
		errorCount := 0
		anomalyCount := 0
		if len(requests) > 0 {
			signal.Add(len(requests))
			for _, request := range requests {
				s.Dispatch("HistCallEthRpc", "rpc_call", request)
			}
			signal.Wait()
			for i, request := range requests {
				blockNumber := request["block_number"].(*big.Int)
				callRpcResult := request.ReturnResult().(*CallEthRpcResult)
				result.Data[i] = &CallEthApiItem{
					ID:    blockNumber.Text(10),
					Data:  string(callRpcResult.Data),
					Error: callRpcResult.Error,
				}
				if result.Data[i].Error != nil {
					errorCount++
					continue
				}
				anomalyCount += s.dispatchTraceTxAnomalies(workerID, blockNumber, callRpcResult.Data)
			}
			if errorCount > 0 {
				result.Error = errors.New("error while invoking api, check inner data for details")
			}
		}
		s.i.Logger.Infof("%s#%02d: %d trace blocks retrieved in %v. Error count = %d. Anomaly count = %d.", s.i.ServiceID, workerID, len(requests), time.Since(startTime), errorCount, anomalyCount)
		msg.Return(result)
	default:
		s.i.Logger.Warnf("%s#%02d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

// Dispatch anomaly records to RecordAnomaly without waiting for the writes.
func (s *CallEthApi) dispatchAnomalyRecords(records []*RecordAnomalyParams) int {
	for _, record := range records {
		s.Dispatch("RecordAnomaly", "record", multiplex.ExecParams{
			"type":         record.Type,
			"block_number": record.BlockNumber,
			"tx_index":     record.TxIndex,
			"tx_hash":      record.TxHash,
			"detail":       record.Detail,
		})
	}
	return len(records)
}

// Detect data anomalies in a successful eth_getBlockByNumber response and dispatch their records to RecordAnomaly.
func (s *CallEthApi) dispatchBlockAnomalies(workerID uint64, blockNumber *big.Int, blockRaw json.RawMessage) int {
	records, err := emptyAttestorRecords(blockNumber, blockRaw)
	if err != nil {
		s.i.Logger.Warnf("%s#%02d: Block anomaly detection failed for block %s. Error: %v.", s.i.ServiceID, workerID, blockNumber.String(), err)
		return 0
	}
	return s.dispatchAnomalyRecords(records)
}

// Detect data anomalies in a successful debug_traceBlockByNumber response and dispatch their records to RecordAnomaly.
func (s *CallEthApi) dispatchTraceTxAnomalies(workerID uint64, blockNumber *big.Int, traceRaw json.RawMessage) int {
	records, err := traceTxErrorRecords(blockNumber, traceRaw)
	if err != nil {
		s.i.Logger.Warnf("%s#%02d: Trace anomaly detection failed for block %s. Error: %v.", s.i.ServiceID, workerID, blockNumber.String(), err)
		return 0
	}
	return s.dispatchAnomalyRecords(records)
}

// CallEthApiParams contains parameters of all commands of CallEthApi.
type CallEthApiParams struct {
	BlockNumbers    []*big.Int
	FromBlockNumber *big.Int
	ToBlockNumber   *big.Int
}

// Extract parameters from a service message and return new CallEthApiParams.
func NewCallEthApiParams(msg *multiplex.ServiceMessage) *CallEthApiParams {
	return &CallEthApiParams{
		BlockNumbers:    msg.GetParam("block_numbers", []*big.Int{}).([]*big.Int),
		FromBlockNumber: msg.GetParam("from_block_number", new(big.Int)).(*big.Int),
		ToBlockNumber:   msg.GetParam("to_block_number", new(big.Int)).(*big.Int),
	}
}

// CallBlockNumberResult contains the result of block_number command.
type CallBlockNumberResult struct {
	Data  uint64
	Error error
}

// CallEthApiResult holds the result of a single call to CallEthRpc.
type CallEthApiItem struct {
	ID    string
	Data  string
	Error error
}

// CallEthApiResult holds the results of all commands of CallEthApi.
type CallEthApiResult struct {
	Method string
	Params []any
	Data   []*CallEthApiItem
	Error  error
}

// Build tracer options for debug_traceBlockByNumber; timeoutSeconds of zero omits the timeout param and keeps the node's default.
func traceBlockTracerOptions(timeoutSeconds uint64) rpc.TraceBlockOptions {
	tracer := "callTracer"
	opts := rpc.TraceBlockOptions{Tracer: &tracer}
	if timeoutSeconds > 0 {
		timeout := fmt.Sprintf("%ds", timeoutSeconds)
		opts.Timeout = &timeout
	}
	return opts
}
