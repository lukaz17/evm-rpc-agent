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
	"testing"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/rs/zerolog"
)

const (
	rpcEndpoint  = "http://eth.tforce.xyz"
	databaseUri  = "mongodb://localhost:27017"
	databaseName = "evm-rpc-agent-test"
)

func newTestController(t *testing.T, dbc *db.DbContext) *Controller {
	t.Helper()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: &nullWriter{}, TimeFormat: time.DateTime}).
		With().
		Timestamp().
		Logger()
	rpcClient := rpc.NewClient(rpcEndpoint)

	cfg := &config.ServiceConfig{
		MaxRpcRrtryCount:         3,
		HistoricalApiWorkerCount: 1,
		StandardApiWorkerCount:   1,
	}
	ctrl := NewController(cfg, rpcClient, dbc, logger)
	go ctrl.Run()
	return ctrl
}
