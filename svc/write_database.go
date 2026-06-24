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

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

// WriteDatabaseResult contains success and failure counts from a database write operation.
type WriteDatabaseResult struct {
	SuccessCount int
	FailedCount  int
}

// WriteDatabase performs write operations to the database.
type WriteDatabase struct {
	multiplex.ServiceCore
	i   *multiplex.ServiceCoreInternal
	dbc *db.DbContext
}

// Return new WriteDatabase instance.
func NewWriteDatabase(dbc *db.DbContext, logger zerolog.Logger) *WriteDatabase {
	svc := &WriteDatabase{
		dbc: dbc,
	}
	svc.i = svc.InitServiceCore("WriteDatabase", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *WriteDatabase) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case db.BlockCollection:
		s.i.Logger.Infof("%s#%d: Block write started.", s.i.ServiceID, workerID)
		p := NewWriteDatabaseParams(msg)
		result := s.writeBlocks(workerID, p.Data)
		msg.Return(result)
	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *WriteDatabase) writeBlocks(workerID uint64, blocks []*CallEthApiItem) *WriteDatabaseResult {
	ctx := context.Background()

	s.i.Logger.Infof("%s#%d: Block write started. Count = %d.", s.i.ServiceID, workerID, len(blocks))

	rawMessages := make([]json.RawMessage, len(blocks))
	for i, block := range blocks {
		rawMessages[i] = json.RawMessage(block.Data)
	}

	err := s.dbc.UpsertRawBlocks(ctx, rawMessages)
	if err != nil {
		s.i.Logger.Warnf("%s#%d: Block write failed. Error: %v", s.i.ServiceID, workerID, err)
		return &WriteDatabaseResult{
			SuccessCount: 0,
			FailedCount:  len(blocks),
		}
	}

	s.i.Logger.Infof("%s#%d: Block write completed. Success = %d.", s.i.ServiceID, workerID, len(blocks))
	return &WriteDatabaseResult{
		SuccessCount: len(blocks),
		FailedCount:  0,
	}
}

// WriteDatabaseParams contains parameters for all commands of WriteDatabase.
type WriteDatabaseParams struct {
	Data []*CallEthApiItem
}

// Get available params from ServiceMessage and return new WriteDatabaseParams.
func NewWriteDatabaseParams(msg *multiplex.ServiceMessage) *WriteDatabaseParams {
	return &WriteDatabaseParams{
		Data: msg.GetParam("data", []*CallEthApiItem{}).([]*CallEthApiItem),
	}
}
