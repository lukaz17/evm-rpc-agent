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
	"fmt"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DefaultTimeout = 10 * time.Second

	BlockCollection = "blocks"
)

// DbContext holds the handles that access MongoDB.
type DbContext struct {
	client *mongo.Client
	db     *mongo.Database
	uri    string
	dbName string
}

// Return a new DbContext instance.
func NewDbContext(cfg *config.DatabaseConfig) *DbContext {
	return &DbContext{
		uri:    cfg.Uri,
		dbName: cfg.Name,
	}
}

// Connect to MongoDB server.
func (dbc *DbContext) Connect() error {
	client, err := mongo.Connect(options.Client().ApplyURI(dbc.uri))
	if err != nil {
		return fmt.Errorf("connect to mongodb: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping mongodb: %w", err)
	}

	dbc.client = client
	dbc.db = client.Database(dbc.dbName)

	return nil
}

// Close connection to MongoDB server.
func (dbc *DbContext) Disconnect() error {
	if dbc.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	return dbc.client.Disconnect(ctx)
}

// Return the underlying mongo.Client.
func (dbc *DbContext) Client() *mongo.Client {
	return dbc.client
}

// Return the underlying mongo.Database.
func (dbc *DbContext) Database() *mongo.Database {
	return dbc.db
}
