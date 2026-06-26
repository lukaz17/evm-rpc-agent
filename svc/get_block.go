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

// GetBlock downloads blockchain data in batches.
type GetBlock struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

// Return new GetBlock instance.
func NewGetBlock(logger zerolog.Logger) *GetBlock {
	svc := &GetBlock{}
	svc.i = svc.InitServiceCore("GetBlock", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *GetBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "get_blocks":
		p := NewGetBlockParams(msg)
		s.downloadBlocks(workerID, p.FromBlockNumber, p.ToBlockNumber, p.BatchSize)
		msg.Return(true)
	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *GetBlock) downloadBlocks(workerID uint64, from, to *big.Int, batch int) {
	s.i.Logger.Infof("%s#%d: Block download started.", s.ServiceID(), workerID)
	batchStartBlockNumber := new(big.Int).Set(from)
	finalBlockNumber := new(big.Int).Set(to)
	for batchStartBlockNumber.Cmp(finalBlockNumber) < 0 {
		batchEndBlockNumber := new(big.Int).Add(batchStartBlockNumber, big.NewInt(int64(batch)-1))
		if batchEndBlockNumber.Cmp(finalBlockNumber) > 0 {
			batchEndBlockNumber.Set(finalBlockNumber)
		}
		getBlocksRequest := multiplex.ExecParams{
			"from_block_number": batchStartBlockNumber,
			"to_block_number":   batchEndBlockNumber,
		}
		getBlocksRequest.ExpectReturn()
		s.Dispatch("CallEthApi", "get_blocks_range", getBlocksRequest)
		getBlocksResponse := getBlocksRequest.WaitForReturn().(*CallEthApiResult)
		getBlockResults := []*CallEthApiItem{}
		for _, blockResult := range getBlocksResponse.Data {
			if blockResult.Error != nil {
				break
			}
			getBlockResults = append(getBlockResults, blockResult)
		}
		lastSuccess := new(big.Int).Add(batchStartBlockNumber, big.NewInt(int64(len(getBlockResults))-1))
		writeBlockRequest := multiplex.ExecParams{
			"data": getBlockResults,
		}
		writeBlockRequest.ExpectReturn()
		s.Dispatch("WriteDatabase", db.BlockCollection, writeBlockRequest)
		writeBlockRequest.Wait()
		result := writeBlockRequest.ReturnResult().(*WriteDatabaseResult)
		if result.SuccessCount > 0 {
			mem.CurrentHeight.Block = lastSuccess.Uint64()
		}
		batchEndBlockNumber.Set(lastSuccess)
		batchStartBlockNumber = new(big.Int).Add(batchEndBlockNumber, big.NewInt(1))
	}
}

// GetBlockParams contain parameters of all commands of GetBlock.
type GetBlockParams struct {
	BatchSize       int
	FromBlockNumber *big.Int
	ToBlockNumber   *big.Int
}

// Extract parameters from a service message and return new GetBlockParams.
func NewGetBlockParams(msg *multiplex.ServiceMessage) *GetBlockParams {
	return &GetBlockParams{
		BatchSize:       msg.GetParam("batch_size", 1).(int),
		FromBlockNumber: msg.GetParam("from_block_number", new(big.Int)).(*big.Int),
		ToBlockNumber:   msg.GetParam("to_block_number", new(big.Int)).(*big.Int),
	}
}
