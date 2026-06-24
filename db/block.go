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

// Block stores structured RPC result of a block in MongoDB.
type Block struct {
	ID        string     `bson:"_id"`
	Number    uint64     `bson:"number"`
	Data      *BlockData `bson:"data"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
}

// BlockData represents an Ethereum block in MongoDB.
type BlockData struct {
	Number           core.Integer   `bson:"number"`
	Time             core.Integer   `bson:"timestamp"`
	Hash             core.Bytes32   `bson:"hash"`
	Nonce            core.Integer   `bson:"nonce"`
	Miner            core.Address   `bson:"miner"`
	Difficulty       core.Integer   `bson:"difficulty"`
	TotalDifficulty  core.Integer   `bson:"totalDifficulty"`
	Size             core.Integer   `bson:"size"`
	GasUsed          core.Integer   `bson:"gasUsed"`
	GasLimit         core.Integer   `bson:"gasLimit"`
	Extra            core.Bytes     `bson:"extraData"`
	MixDigest        core.Bytes32   `bson:"mixHash"`
	LogsBloom        core.Bytes     `bson:"logsBloom"`
	ParentHash       core.Bytes32   `bson:"parentHash"`
	UncleHash        core.Bytes32   `bson:"sha3Uncles"`
	TransactionsRoot core.Bytes32   `bson:"transactionsRoot"`
	ReceiptsRoot     core.Bytes32   `bson:"receiptsRoot"`
	StateRoot        core.Bytes32   `bson:"stateRoot"`
	Transactions     []Transaction  `bson:"transactions,omitempty"`
	Uncles           []core.Bytes32 `bson:"uncles,omitempty"`
}

// Transaction represents an Ethereum transaction in MongoDB.
type Transaction struct {
	Hash        core.Bytes32 `bson:"hash"`
	BlockNumber core.Integer `bson:"blockNumber"`
	BlockHash   core.Bytes32 `bson:"blockHash"`
	BlockTime   core.Integer `bson:"blockTimestamp"`
	Nonce       core.Integer `bson:"nonce"`
	Index       core.Integer `bson:"transactionIndex"`
	From        core.Address `bson:"from"`
	To          core.Address `bson:"to"`
	Value       core.Integer `bson:"value"`
	Gas         core.Integer `bson:"gas"`
	GasPrice    core.Integer `bson:"gasPrice"`
	Type        core.Integer `bson:"type"`
	Input       core.Bytes   `bson:"input"`
	V           core.Integer `bson:"v"`
	R           core.Bytes   `bson:"r"`
	S           core.Bytes   `bson:"s"`
}

// Return BSON Block from RPC Block.
func NewBlockFromRPC(rpcBlock *rpc.Block) (*Block, error) {
	if rpcBlock == nil {
		return nil, nil
	}

	blockNumber := rpcBlock.Number.Uint64()
	data := &BlockData{
		Number:           rpcBlock.Number,
		Time:             rpcBlock.Time,
		Hash:             rpcBlock.Hash,
		Nonce:            rpcBlock.Nonce,
		Miner:            rpcBlock.Coinbase,
		Difficulty:       rpcBlock.Difficulty,
		TotalDifficulty:  rpcBlock.TotalDifficulty,
		Size:             rpcBlock.Size,
		GasUsed:          rpcBlock.GasUsed,
		GasLimit:         rpcBlock.GasLimit,
		Extra:            rpcBlock.Extra,
		MixDigest:        rpcBlock.MixDigest,
		LogsBloom:        rpcBlock.LogsBloom,
		ParentHash:       rpcBlock.ParentHash,
		UncleHash:        rpcBlock.UncleHash,
		TransactionsRoot: rpcBlock.TransactionsRoot,
		ReceiptsRoot:     rpcBlock.ReceiptsRoot,
		StateRoot:        rpcBlock.StateRoot,
	}

	txs, err := rpcBlock.TransactionsFull()
	if err == nil && len(txs) > 0 {
		data.Transactions = make([]Transaction, len(txs))
		for i, tx := range txs {
			data.Transactions[i] = NewTransactionFromRPC(tx)
		}
	} else {
		hashes, err := rpcBlock.TransactionsHashes()
		if err == nil && len(hashes) > 0 {
			data.Transactions = make([]Transaction, len(hashes))
			for i, h := range hashes {
				data.Transactions[i] = Transaction{Hash: h}
			}
		}
	}

	uncles, err := rpcBlock.Uncles()
	if err == nil && len(uncles) > 0 {
		data.Uncles = uncles
	}

	utcNow := time.Now().UTC()
	return &Block{
		ID:        fmt.Sprintf("%d", blockNumber),
		Number:    blockNumber,
		Data:      data,
		CreatedAt: utcNow,
		UpdatedAt: utcNow,
	}, nil
}

// Return BSON Transaction from RPC Transaction.
func NewTransactionFromRPC(tx rpc.Transaction) Transaction {
	return Transaction{
		Hash:        tx.Hash,
		BlockNumber: tx.BlockNumber,
		BlockHash:   tx.BlockHash,
		BlockTime:   tx.BlockTime,
		Nonce:       tx.Nonce,
		Index:       tx.Index,
		From:        tx.From,
		To:          tx.To,
		Value:       tx.Value,
		Gas:         tx.Gas,
		GasPrice:    tx.GasPrice,
		Type:        tx.Type,
		Input:       tx.Input,
		V:           tx.V,
		R:           tx.R,
		S:           tx.S,
	}
}

// Get a block by ID.
func (dbc *DbContext) GetBlock(ctx context.Context, number uint64) (*Block, error) {
	coll := dbc.db.Collection(BlockCollection)
	var block Block
	err := coll.FindOne(ctx, bson.M{"_id": fmt.Sprintf("%d", number)}).Decode(&block)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &block, nil
}

// Get a block by ID without ORM mapping.
func (dbc *DbContext) GetRawBlock(ctx context.Context, number uint64) (*RawBlockData, error) {
	coll := dbc.db.Collection(BlockCollection)
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

// Insert a new Block document.
func (dbc *DbContext) InsertBlock(ctx context.Context, block *Block) error {
	return dbc.insertBlock(ctx, block, false)
}

// Insert multiple Block documents.
func (dbc *DbContext) InsertBlocks(ctx context.Context, blocks []*Block) error {
	if len(blocks) == 0 {
		return nil
	}

	coll := dbc.db.Collection(BlockCollection)
	utcNow := time.Now().UTC()

	docs := make([]interface{}, len(blocks))
	for i, block := range blocks {
		block.ID = fmt.Sprintf("%d", block.Number)
		if block.CreatedAt.IsZero() {
			block.CreatedAt = utcNow
		}
		block.UpdatedAt = utcNow
		docs[i] = block
	}

	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		if result != nil && len(result.InsertedIDs) > 0 {
			insertedIDs := make([]string, 0, len(result.InsertedIDs))
			for _, id := range result.InsertedIDs {
				if s, ok := id.(string); ok {
					insertedIDs = append(insertedIDs, s)
				}
			}
			if len(insertedIDs) > 0 {
				coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": insertedIDs}})
			}
		}
		return fmt.Errorf("InsertBlocks: %w", err)
	}
	return nil
}

// Insert a new Block document without ORM mapping.
func (dbc *DbContext) InsertRawBlock(ctx context.Context, raw json.RawMessage) error {
	return dbc.insertRawBlock(ctx, raw, false)
}

// Insert multiple Block documents without ORM mapping.
func (dbc *DbContext) InsertRawBlocks(ctx context.Context, raws []json.RawMessage) error {
	if len(raws) == 0 {
		return nil
	}

	coll := dbc.db.Collection(BlockCollection)
	utcNow := time.Now().UTC()

	docs := make([]*RawBlockData, len(raws))
	for i, raw := range raws {
		data, err := JsonToBson(raw)
		if err != nil {
			return fmt.Errorf("InsertRawBlocks: failed to convert JSON to BSON at index %d: %w", i, err)
		}

		blockNumber, err := getNumberFromDoc(data, "number")
		if err != nil {
			return fmt.Errorf("InsertRawBlocks: block number not found at index %d: %w", i, err)
		}

		id := fmt.Sprintf("%d", blockNumber)
		docs[i] = &RawBlockData{
			ID:        id,
			Number:    blockNumber,
			Data:      data,
			CreatedAt: utcNow,
			UpdatedAt: utcNow,
		}
	}

	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		if result != nil && len(result.InsertedIDs) > 0 {
			insertedIDs := make([]string, 0, len(result.InsertedIDs))
			for _, id := range result.InsertedIDs {
				if s, ok := id.(string); ok {
					insertedIDs = append(insertedIDs, s)
				}
			}
			if len(insertedIDs) > 0 {
				coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": insertedIDs}})
			}
		}
		return fmt.Errorf("InsertRawBlocks: %w", err)
	}
	return nil
}

// Insert or replace a Block document.
func (dbc *DbContext) UpsertBlock(ctx context.Context, block *Block) error {
	return dbc.insertBlock(ctx, block, true)
}

// Insert or replace multiple Block documents.
func (dbc *DbContext) UpsertBlocks(ctx context.Context, blocks []*Block) error {
	if len(blocks) == 0 {
		return nil
	}

	coll := dbc.db.Collection(BlockCollection)
	utcNow := time.Now().UTC()

	models := make([]mongo.WriteModel, len(blocks))
	for i, block := range blocks {
		block.ID = fmt.Sprintf("%d", block.Number)
		if block.CreatedAt.IsZero() {
			block.CreatedAt = utcNow
		}
		block.UpdatedAt = utcNow

		filter := bson.M{"_id": block.ID}
		models[i] = mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(block).SetUpsert(true)
	}

	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		if result != nil && len(result.UpsertedIDs) > 0 {
			upsertedIDs := make([]string, 0, len(result.UpsertedIDs))
			for _, id := range result.UpsertedIDs {
				if s, ok := id.(string); ok {
					upsertedIDs = append(upsertedIDs, s)
				}
			}
			if len(upsertedIDs) > 0 {
				coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": upsertedIDs}})
			}
		}
		return fmt.Errorf("UpsertBlocks: %w", err)
	}
	return nil
}

// Insert or replace a Block document without ORM mapping.
func (dbc *DbContext) UpsertRawBlock(ctx context.Context, raw json.RawMessage) error {
	return dbc.insertRawBlock(ctx, raw, true)
}

// Insert or replace multiple Block documents without ORM mapping.
func (dbc *DbContext) UpsertRawBlocks(ctx context.Context, raws []json.RawMessage) error {
	if len(raws) == 0 {
		return nil
	}

	coll := dbc.db.Collection(BlockCollection)
	utcNow := time.Now().UTC()

	models := make([]mongo.WriteModel, len(raws))
	for i, raw := range raws {
		data, err := JsonToBson(raw)
		if err != nil {
			return fmt.Errorf("UpsertRawBlocks: failed to convert JSON to BSON at index %d: %w", i, err)
		}

		blockNumber, err := getNumberFromDoc(data, "number")
		if err != nil {
			return fmt.Errorf("UpsertRawBlocks: block number not found at index %d: %w", i, err)
		}

		id := fmt.Sprintf("%d", blockNumber)
		doc := &RawBlockData{
			ID:        id,
			Number:    blockNumber,
			Data:      data,
			CreatedAt: utcNow,
			UpdatedAt: utcNow,
		}

		filter := bson.M{"_id": id}
		models[i] = mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(doc).SetUpsert(true)
	}

	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		if result != nil && len(result.UpsertedIDs) > 0 {
			upsertedIDs := make([]string, 0, len(result.UpsertedIDs))
			for _, id := range result.UpsertedIDs {
				if s, ok := id.(string); ok {
					upsertedIDs = append(upsertedIDs, s)
				}
			}
			if len(upsertedIDs) > 0 {
				coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": upsertedIDs}})
			}
		}
		return fmt.Errorf("UpsertRawBlocks: %w", err)
	}
	return nil
}

// Delete a block by ID
func (dbc *DbContext) DeleteBlock(ctx context.Context, number uint64) error {
	coll := dbc.db.Collection(BlockCollection)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": fmt.Sprintf("%d", number)})
	return err
}

// Return size of Block collection.
func (dbc *DbContext) CountBlocks(ctx context.Context) (int64, error) {
	coll := dbc.db.Collection(BlockCollection)
	return coll.CountDocuments(ctx, bson.M{})
}

func (dbc *DbContext) insertBlock(ctx context.Context, block *Block, upsert bool) error {
	coll := dbc.db.Collection(BlockCollection)
	utcNow := time.Now().UTC()

	block.ID = fmt.Sprintf("%d", block.Number)
	if block.CreatedAt.IsZero() {
		block.CreatedAt = utcNow
	}
	block.UpdatedAt = utcNow

	if upsert {
		filter := bson.M{"_id": block.ID}
		opts := options.Replace().SetUpsert(true)
		_, err := coll.ReplaceOne(ctx, filter, block, opts)
		return err
	}

	_, err := coll.InsertOne(ctx, block)
	return err
}

func (dbc *DbContext) insertRawBlock(ctx context.Context, raw json.RawMessage, upsert bool) error {
	coll := dbc.db.Collection(BlockCollection)

	data, err := JsonToBson(raw)
	if err != nil {
		return fmt.Errorf("insertRawBlock: failed to convert JSON to BSON: %w", err)
	}

	blockNumer, err := getNumberFromDoc(data, "number")
	if err != nil {
		return fmt.Errorf("insertRawBlock: block number not found")
	}

	utcNow := time.Now().UTC()
	id := fmt.Sprintf("%d", blockNumer)
	doc := &RawBlockData{
		ID:        id,
		Number:    blockNumer,
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
