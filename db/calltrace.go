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

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lukaz17/evm-rpc-agent/core"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BlockCallTrace stores all call traces of a block stored in MongoDB.
type BlockCallTrace struct {
	ID        string                 `bson:"_id"`
	Number    uint64                 `bson:"number"`
	Data      []TransactionCallTrace `bson:"data"`
	CreatedAt time.Time              `bson:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at"`
}

// CallFrame represents a single call frame from callTracer stored in MongoDB.
type CallFrame struct {
	From    core.Address `bson:"from"`
	To      core.Address `bson:"to"`
	Gas     core.Integer `bson:"gas"`
	GasUsed core.Integer `bson:"gasUsed"`
	Value   core.Integer `bson:"value"`
	Type    string       `bson:"type"`
	Input   core.Bytes   `bson:"input"`
	Output  core.Bytes   `bson:"output"`
	Calls   []CallFrame  `bson:"calls"`
}

// TransactionCallTrace represents a single transaction call trace within a block.
type TransactionCallTrace struct {
	TxHash string    `bson:"txHash"`
	Result CallFrame `bson:"result"`
}

// Return a BSON BlockCallTrace from RPC TransactionTraces.
func NewBlockCallTraceFromRPC(blockNum uint64, rpcTraces []rpc.TransactionTrace) (*BlockCallTrace, error) {
	if len(rpcTraces) == 0 {
		return nil, nil
	}

	traces := make([]TransactionCallTrace, 0, len(rpcTraces))

	for _, rt := range rpcTraces {
		cf, err := rt.CallTraceResult()
		if err != nil {
			return nil, fmt.Errorf("extract call frame for tx %s: %w", rt.TxHash.Hex(), err)
		}

		traces = append(traces, TransactionCallTrace{
			TxHash: rt.TxHash.Hex(),
			Result: newCallFrameFromRPC(*cf),
		})
	}

	return &BlockCallTrace{
		ID:        fmt.Sprintf("%d", blockNum),
		Number:    blockNum,
		Data:      traces,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// newCallFrameFromRPC converts an RPC CallFrame to a BSON CallFrame.
func newCallFrameFromRPC(cf rpc.CallFrame) CallFrame {
	return CallFrame{
		From:    cf.From,
		To:      cf.To,
		Gas:     cf.Gas,
		GasUsed: cf.GasUsed,
		Value:   cf.Value,
		Type:    cf.Type,
		Input:   cf.Input,
		Output:  cf.Output,
		Calls:   newCallFramesFromRPC(cf.Calls),
	}
}

func newCallFramesFromRPC(cfs []rpc.CallFrame) []CallFrame {
	if len(cfs) == 0 {
		return nil
	}
	out := make([]CallFrame, len(cfs))
	for i, c := range cfs {
		out[i] = newCallFrameFromRPC(c)
	}
	return out
}

// Return a call trace by ID.
func (dbc *DbContext) GetCallTrace(ctx context.Context, number uint64) (*BlockCallTrace, error) {
	coll := dbc.db.Collection(CallTraceCollection)
	var trace BlockCallTrace
	err := coll.FindOne(ctx, bson.M{"_id": fmt.Sprintf("%d", number)}).Decode(&trace)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &trace, nil
}

// Return a call trace by ID without ORM mapping.
func (dbc *DbContext) GetRawCallTrace(ctx context.Context, number uint64) (*RawBlockData, error) {
	coll := dbc.db.Collection(CallTraceCollection)
	var data RawBlockData
	err := coll.FindOne(ctx, bson.M{"_id": fmt.Sprintf("%d", number)}).Decode(&data)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

// Insert a new BlockCallTrace document.
func (dbc *DbContext) InsertCallTrace(ctx context.Context, trace *BlockCallTrace) error {
	return dbc.insertCallTrace(ctx, trace, false)
}

// Insert a new BlockCallTrace document without ORM mapping.
func (dbc *DbContext) InsertRawCallTrace(ctx context.Context, blockNumber uint64, raw json.RawMessage) error {
	return dbc.insertRawCallTrace(ctx, blockNumber, raw, false)
}

// Insert or replace a BlockCallTrace document.
func (dbc *DbContext) UpsertCallTrace(ctx context.Context, trace *BlockCallTrace) error {
	return dbc.insertCallTrace(ctx, trace, true)
}

// Insert or replace a BlockCallTrace document without ORM mapping.
func (dbc *DbContext) UpsertRawCallTrace(ctx context.Context, blockNumber uint64, raw json.RawMessage) error {
	return dbc.insertRawCallTrace(ctx, blockNumber, raw, true)
}

// Delete a call trace by ID
func (dbc *DbContext) DeleteCallTrace(ctx context.Context, number uint64) error {
	coll := dbc.db.Collection(CallTraceCollection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": fmt.Sprintf("%d", number)})
	return err
}

// Return size of CallTrace collection.
func (dbc *DbContext) CountCallTraces(ctx context.Context) (int64, error) {
	coll := dbc.db.Collection(CallTraceCollection)
	return coll.CountDocuments(ctx, bson.M{})
}

func (dbc *DbContext) insertCallTrace(ctx context.Context, trace *BlockCallTrace, upsert bool) error {
	coll := dbc.db.Collection(CallTraceCollection)
	utcNow := time.Now().UTC()

	trace.ID = fmt.Sprintf("%d", trace.Number)
	if trace.CreatedAt.IsZero() {
		trace.CreatedAt = utcNow
	}
	trace.UpdatedAt = utcNow

	if upsert {
		filter := bson.M{"_id": trace.ID}
		opts := options.Replace().SetUpsert(true)
		_, err := coll.ReplaceOne(ctx, filter, trace, opts)
		return err
	}

	_, err := coll.InsertOne(ctx, trace)
	return err
}

func (dbc *DbContext) insertRawCallTrace(ctx context.Context, blockNumber uint64, raw json.RawMessage, upsert bool) error {
	coll := dbc.db.Collection(CallTraceCollection)

	data, err := JsonToBsonArr(raw)
	if err != nil {
		return fmt.Errorf("insertRawBlock: failed to convert JSON to BSON: %w", err)
	}

	utcNow := time.Now().UTC()
	id := fmt.Sprintf("%d", blockNumber)
	doc := &RawBlockArray{
		ID:        id,
		Number:    blockNumber,
		Data:      data,
		CreatedAt: utcNow,
		UpdatedAt: utcNow,
	}

	if upsert {
		filter := bson.M{"_id": id}
		opts := options.Replace().SetUpsert(true)
		_, err = coll.ReplaceOne(ctx, filter, doc, opts)
		return err
	}

	_, err = coll.InsertOne(ctx, doc)
	return err
}
