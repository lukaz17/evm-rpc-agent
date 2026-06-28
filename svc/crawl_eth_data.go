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
	"github.com/lukaz17/evm-rpc-agent/mem"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

// CrawlEthData continuously crawls new blocks from the chain.
type CrawlEthData struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal

	batchSize  uint64
	blockDelay uint64
	target     *LatestBlock
}

// Return new CrawlEthData instance.
func NewCrawlEthData(batchSize uint64, blockDelay uint64, logger zerolog.Logger) *CrawlEthData {
	svc := &CrawlEthData{
		batchSize:  batchSize,
		blockDelay: blockDelay,
		target: &LatestBlock{
			Block:     new(multiplex.Uint64ThreadSafe),
			CallTrace: new(multiplex.Uint64ThreadSafe),
		},
	}
	svc.i = svc.InitServiceCore("CrawlEthData", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *CrawlEthData) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "crawl":
		if !s.target.Initialized {
			s.target.Initialized = true
			s.target.Block.Set(mem.CurrentHeight.Block)
			s.target.CallTrace.Set(mem.CurrentHeight.CallTrace)
		}

		blockNumberMsg := multiplex.ExecParams{}
		blockNumberMsg.ExpectReturn()
		s.Dispatch("CallEthApi", "block_number", blockNumberMsg)
		blockNumResult := blockNumberMsg.WaitForReturn().(*CallBlockNumberResult)
		if blockNumResult.Error != nil {
			s.i.Logger.Warnf("%s#%d: Failed to get latest block number: %v.", s.i.ServiceID, workerID, blockNumResult.Error)
			msg.Return(false)
			return &multiplex.HookState{Handled: true}
		}

		safeBlock := blockNumResult.Data
		if s.blockDelay >= safeBlock {
			safeBlock = 0
		} else {
			safeBlock -= s.blockDelay
		}

		s.dispatchBlocks(workerID, safeBlock)
		s.dispatchCallTraces(workerID, safeBlock)

		msg.Return(true)

	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *CrawlEthData) dispatchBlocks(workerID uint64, latestBlock uint64) {
	s.target.Block.Lock()
	defer s.target.Block.Unlock()

	fromBlock := mem.CurrentHeight.Block
	targetVal := s.target.Block.ValueNoLock()
	if fromBlock < targetVal {
		s.i.Logger.Warnf("%s#%d: Crawl already in progress (cached=%d, target=%d), skipping.", s.i.ServiceID, workerID, fromBlock, targetVal)
		return
	}
	if fromBlock >= latestBlock {
		s.i.Logger.Infof("%s#%d: Up to date (cached=%d, latest=%d).", s.i.ServiceID, workerID, fromBlock, latestBlock)
		return
	}

	mem.TargetHeight.Block = latestBlock
	s.target.Block.SetNoLock(latestBlock)
	s.i.Logger.Infof("%s#%d: Fetching blocks %d to %d.", s.i.ServiceID, workerID, fromBlock+1, latestBlock)
	getBlocksReq := multiplex.ExecParams{
		"from_block_number": new(big.Int).SetUint64(fromBlock + 1),
		"to_block_number":   new(big.Int).SetUint64(latestBlock),
		"batch_size":        s.batchSize,
	}
	s.Dispatch("GetBlock", "get_blocks", getBlocksReq)
}

func (s *CrawlEthData) dispatchCallTraces(workerID uint64, latestBlock uint64) {
	s.target.CallTrace.Lock()
	defer s.target.CallTrace.Unlock()

	fromBlock := mem.CurrentHeight.CallTrace
	targetVal := s.target.CallTrace.ValueNoLock()
	if fromBlock < targetVal {
		s.i.Logger.Warnf("%s#%d: Call Trace crawl already in progress (cached=%d, target=%d), skipping.", s.i.ServiceID, workerID, fromBlock, targetVal)
		return
	}
	if fromBlock >= latestBlock {
		s.i.Logger.Infof("%s#%d: Call Traces up to date (cached=%d, latest=%d).", s.i.ServiceID, workerID, fromBlock, latestBlock)
		return
	}

	mem.TargetHeight.CallTrace = latestBlock
	s.target.CallTrace.SetNoLock(latestBlock)
	s.i.Logger.Infof("%s#%d: Fetching call traces %d to %d.", s.i.ServiceID, workerID, fromBlock+1, latestBlock)
	getTracesReq := multiplex.ExecParams{
		"from_block_number": new(big.Int).SetUint64(fromBlock + 1),
		"to_block_number":   new(big.Int).SetUint64(latestBlock),
		"batch_size":        s.batchSize,
	}
	s.Dispatch("TraceBlock", "trace_calls", getTracesReq)
}

type LatestBlock struct {
	Block     *multiplex.Uint64ThreadSafe
	CallTrace *multiplex.Uint64ThreadSafe

	Initialized bool
}
