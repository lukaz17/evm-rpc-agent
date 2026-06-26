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
	"math/big"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/lukaz17/evm-rpc-agent/mem"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

// TraceBlock downloads blockchain trace data in batches.
type TraceBlock struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

// Return new TraceBlock instance.
func NewTraceBlock(logger zerolog.Logger) *TraceBlock {
	svc := &TraceBlock{}
	svc.i = svc.InitServiceCore("TraceBlock", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *TraceBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "trace_calls":
		p := NewTraceBlockParams(msg)
		s.downloadTraceBlocks(workerID, p.FromBlockNumber, p.ToBlockNumber, p.BatchSize)
		msg.Return(true)
	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *TraceBlock) downloadTraceBlocks(workerID uint64, from, to *big.Int, batch int) {
	s.i.Logger.Infof("%s#%d: Trace block download started.", s.ServiceID(), workerID)
	batchStartBlockNumber := new(big.Int).Set(from)
	finalBlockNumber := new(big.Int).Set(to)
	for batchStartBlockNumber.Cmp(finalBlockNumber) <= 0 {
		batchEndBlockNumber := new(big.Int).Add(batchStartBlockNumber, big.NewInt(int64(batch)-1))
		if batchEndBlockNumber.Cmp(finalBlockNumber) > 0 {
			batchEndBlockNumber.Set(finalBlockNumber)
		}
		traceBlocksRequest := multiplex.ExecParams{
			"from_block_number": batchStartBlockNumber,
			"to_block_number":   batchEndBlockNumber,
		}
		traceBlocksRequest.ExpectReturn()
		s.Dispatch("CallEthApi", "trace_calls_range", traceBlocksRequest)
		traceBlocksResponse := traceBlocksRequest.WaitForReturn().(*CallEthApiResult)
		getTraceResults := []*CallEthApiItem{}
		for _, traceResult := range traceBlocksResponse.Data {
			if traceResult.Error != nil {
				break
			}
			getTraceResults = append(getTraceResults, traceResult)
		}
		lastSuccess := new(big.Int).Add(batchStartBlockNumber, big.NewInt(int64(len(getTraceResults))-1))
		writeTraceRequest := multiplex.ExecParams{
			"data": getTraceResults,
		}
		writeTraceRequest.ExpectReturn()
		s.Dispatch("WriteDatabase", db.CallTraceCollection, writeTraceRequest)
		writeTraceRequest.Wait()
		result := writeTraceRequest.ReturnResult().(*WriteDatabaseResult)
		if result.SuccessCount > 0 {
			mem.CurrentHeight.CallTrace = lastSuccess.Uint64()
		}
		batchEndBlockNumber.Set(lastSuccess)
		batchStartBlockNumber = new(big.Int).Add(batchEndBlockNumber, big.NewInt(1))
	}
}

// TraceBlockParams contain parameters of all commands of TraceBlock.
type TraceBlockParams struct {
	BatchSize       int
	FromBlockNumber *big.Int
	ToBlockNumber   *big.Int
}

// Extract parameters from a service message and return new TraceBlockParams.
func NewTraceBlockParams(msg *multiplex.ServiceMessage) *TraceBlockParams {
	return &TraceBlockParams{
		BatchSize:       msg.GetParam("batch_size", 1).(int),
		FromBlockNumber: msg.GetParam("from_block_number", new(big.Int)).(*big.Int),
		ToBlockNumber:   msg.GetParam("to_block_number", new(big.Int)).(*big.Int),
	}
}
