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

// BlockData represents an Ethereum block stored in MongoDB.
type BlockData struct {
	Number           uint64        `bson:"number"`
	Time             uint64        `bson:"timestamp"`
	Hash             string        `bson:"hash"`
	Nonce            string        `bson:"nonce"`
	Miner            string        `bson:"miner"`
	Difficulty       string        `bson:"difficulty"`
	TotalDifficulty  string        `bson:"totalDifficulty"`
	Size             uint64        `bson:"size"`
	GasUsed          uint64        `bson:"gasUsed"`
	GasLimit         uint64        `bson:"gasLimit"`
	Extra            string        `bson:"extraData"`
	MixDigest        string        `bson:"mixHash"`
	LogsBloom        string        `bson:"logsBloom"`
	ParentHash       string        `bson:"parentHash"`
	UncleHash        string        `bson:"sha3Uncles"`
	TransactionsRoot string        `bson:"transactionsRoot"`
	ReceiptsRoot     string        `bson:"receiptsRoot"`
	StateRoot        string        `bson:"stateRoot"`
	Transactions     []Transaction `bson:"transactions,omitempty"`
	Uncles           []string      `bson:"uncles,omitempty"`
}

// Transaction represents an Ethereum transaction stored in MongoDB.
type Transaction struct {
	Hash        string `bson:"hash"`
	BlockNumber uint64 `bson:"blockNumber"`
	BlockHash   string `bson:"blockHash"`
	BlockTime   uint64 `bson:"blockTimestamp"`
	Nonce       string `bson:"nonce"`
	Index       uint64 `bson:"transactionIndex"`
	From        string `bson:"from"`
	To          string `bson:"to"`
	Value       string `bson:"value"`
	Gas         uint64 `bson:"gas"`
	GasPrice    string `bson:"gasPrice"`
	Type        uint64 `bson:"type"`
	Input       string `bson:"input"`
	V           string `bson:"v"`
	R           string `bson:"r"`
	S           string `bson:"s"`
}

// Return a BSON Block from RPC Block.
func NewBlockFromRPC(rpcBlock *rpc.Block) (*Block, error) {
	if rpcBlock == nil {
		return nil, nil
	}

	blockNumber := rpcBlock.Number.Uint64()
	data := &BlockData{
		Number:           blockNumber,
		Time:             rpcBlock.Time.Uint64(),
		Hash:             rpcBlock.Hash.Hex(),
		Nonce:            rpcBlock.Nonce.Hex(),
		Miner:            rpcBlock.Coinbase.Hex(),
		Difficulty:       rpcBlock.Difficulty.Hex(),
		TotalDifficulty:  rpcBlock.TotalDifficulty.Hex(),
		Size:             rpcBlock.Size.Uint64(),
		GasUsed:          rpcBlock.GasUsed.Uint64(),
		GasLimit:         rpcBlock.GasLimit.Uint64(),
		Extra:            rpcBlock.Extra.Hex(),
		MixDigest:        rpcBlock.MixDigest.Hex(),
		LogsBloom:        rpcBlock.LogsBloom.Hex(),
		ParentHash:       rpcBlock.ParentHash.Hex(),
		UncleHash:        rpcBlock.UncleHash.Hex(),
		TransactionsRoot: rpcBlock.TransactionsRoot.Hex(),
		ReceiptsRoot:     rpcBlock.ReceiptsRoot.Hex(),
		StateRoot:        rpcBlock.StateRoot.Hex(),
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
				data.Transactions[i] = Transaction{Hash: h.Hex()}
			}
		}
	}

	uncles, err := rpcBlock.Uncles()
	if err == nil && len(uncles) > 0 {
		data.Uncles = make([]string, len(uncles))
		for i, u := range uncles {
			data.Uncles[i] = u.Hex()
		}
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

// Return a BSON Transaction from RPC Transaction.
func NewTransactionFromRPC(tx rpc.Transaction) Transaction {
	return Transaction{
		Hash:        tx.Hash.Hex(),
		BlockNumber: tx.BlockNumber.Uint64(),
		BlockHash:   tx.BlockHash.Hex(),
		BlockTime:   tx.BlockTime.Uint64(),
		Nonce:       tx.Nonce.Hex(),
		Index:       tx.Index.Uint64(),
		From:        tx.From.Hex(),
		To:          tx.To.Hex(),
		Value:       tx.Value.Hex(),
		Gas:         tx.Gas.Uint64(),
		GasPrice:    tx.GasPrice.Hex(),
		Type:        tx.Type.Uint64(),
		Input:       tx.Input.Hex(),
		V:           tx.V.Hex(),
		R:           tx.R.Hex(),
		S:           tx.S.Hex(),
	}
}

// Return a block by ID.
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

// Return a block by ID without ORM mapping.
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

// Insert a new Block document without ORM mapping.
func (dbc *DbContext) InsertRawBlock(ctx context.Context, raw json.RawMessage) error {
	return dbc.insertRawBlock(ctx, raw, false)
}

// Insert or replace a Block document.
func (dbc *DbContext) UpsertBlock(ctx context.Context, block *Block) error {
	return dbc.insertBlock(ctx, block, true)
}

// Insert or replace a Block document without ORM mapping.
func (dbc *DbContext) UpsertRawBlock(ctx context.Context, raw json.RawMessage) error {
	return dbc.insertRawBlock(ctx, raw, true)
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
