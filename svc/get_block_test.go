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
	"github.com/tforce-io/tf-golib/multiplex"
)

func TestGetBlock_DownloadBlocks(t *testing.T) {
	fromBlock := big.NewInt(102)
	toBlock := big.NewInt(117)
	batchSize := 8
	expectedCount := int(toBlock.Int64()-fromBlock.Int64()) + 1

	dbc := newTestDbContext(t, []string{db.BlockCollection})
	ctrl := newTestController(t, dbc)

	params := multiplex.ExecParams{
		"from_block_number": fromBlock,
		"to_block_number":   toBlock,
		"batch_size":        batchSize,
	}
	ctrl.DispatchOnce("GetBlock", "get_blocks", params)

	result := params.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}

	time.Sleep(500 * time.Millisecond)

	finalCount, err := dbc.CountBlocks(context.Background())
	if err != nil {
		t.Fatalf("count blocks: %v", err)
	}

	if int(finalCount) < expectedCount {
		t.Errorf("expected at least %d new blocks in DB, got %d", expectedCount, finalCount)
	}

	for blockNum := fromBlock.Int64(); blockNum <= toBlock.Int64(); blockNum++ {
		block, err := dbc.GetBlock(context.Background(), uint64(blockNum))
		if err != nil {
			t.Errorf("get block %d: %v", blockNum, err)
			continue
		}

		if block == nil {
			t.Errorf("block %d not found in database", blockNum)
			continue
		}

		if block.Number != uint64(blockNum) {
			t.Errorf("block number = %d, want %d", block.Number, blockNum)
		}
	}
}
