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
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/tforce-io/tf-golib/multiplex"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newTestDbContext(t *testing.T, collNames []string) *db.DbContext {
	t.Helper()
	dbc := db.NewDbContext(&config.DatabaseConfig{
		Uri:  databaseUri,
		Name: databaseName,
	})
	if err := dbc.Connect(); err != nil {
		t.Fatalf("connect to mongodb: %v", err)
	}

	for _, collName := range collNames {
		coll := dbc.Database().Collection(collName)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if _, err := coll.DeleteMany(ctx, bson.M{}); err != nil {
			t.Fatalf("clean collection %s: %v", collName, err)
		}
		cancel()

		collName := collName
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if _, err := dbc.Database().Collection(collName).DeleteMany(ctx, bson.M{}); err != nil {
				t.Logf("cleanup collection %s: %v", collName, err)
			}
		})
	}

	t.Cleanup(func() {
		if err := dbc.Disconnect(); err != nil {
			t.Logf("disconnect failed: %v", err)
		}
	})

	return dbc
}

func TestWriteDatabase_Blocks(t *testing.T) {
	blobData, err := os.ReadFile("../rpc/blobs/blocks.json")
	if err != nil {
		t.Fatalf("read blocks.json: %v", err)
	}

	var blocksMap map[string]json.RawMessage
	if err := json.Unmarshal(blobData, &blocksMap); err != nil {
		t.Fatalf("unmarshal blocks.json: %v", err)
	}

	items := make([]*CallEthApiItem, 0, len(blocksMap))
	for hexNum, rawBlock := range blocksMap {
		num, _ := new(big.Int).SetString(hexNum[2:], 16)
		items = append(items, &CallEthApiItem{
			ID:   num.Text(10),
			Data: string(rawBlock),
		})
	}

	dbc := newTestDbContext(t, []string{db.BlockCollection})
	ctrl := newTestController(t, dbc)

	params := multiplex.ExecParams{
		"data": items,
	}
	ctrl.DispatchOnce("WriteDatabase", db.BlockCollection, params)

	result, ok := params.ReturnResult().(*WriteDatabaseResult)
	if !ok {
		t.Fatalf("result is not *WriteDatabaseResult, got %T", result)
	}

	if result.SuccessCount != len(items) {
		t.Errorf("success count = %d, want %d", result.SuccessCount, len(items))
	}

	if result.FailedCount != 0 {
		t.Errorf("failed count = %d, want 0", result.FailedCount)
	}

	finalCount, err := dbc.CountBlocks(context.Background())
	if err != nil {
		t.Fatalf("count blocks: %v", err)
	}

	if int(finalCount) != len(items) {
		t.Errorf("expected %d blocks in DB, got %d", len(items), finalCount)
	}

	for hexNum := range blocksMap {
		num, _ := new(big.Int).SetString(hexNum[2:], 16)
		block, err := dbc.GetBlock(context.Background(), num.Uint64())
		if err != nil {
			t.Errorf("get block %s: %v", hexNum, err)
			continue
		}

		if block == nil {
			t.Errorf("block %s not found in database", hexNum)
			continue
		}

		if block.Number != num.Uint64() {
			t.Errorf("block number = %d, want %d", block.Number, num.Uint64())
		}
	}
}
