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
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/mem"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

// Notify sends notifications to users.
type Notify struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

// Return new Notify instance.
func NewNotify(logger zerolog.Logger) *Notify {
	svc := &Notify{}
	svc.i = svc.InitServiceCore("Notify", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *Notify) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "height_status":
		s.i.Logger.Infof("CurrentHeight -> Block: %d, CallTrace: %d | TargetHeight -> Block: %d, CallTrace: %d",
			mem.CurrentHeight.Block, mem.CurrentHeight.CallTrace,
			mem.TargetHeight.Block, mem.TargetHeight.CallTrace)
		msg.Return(true)
	case "last_update":
		mem.LastUpdate = time.Now().UnixMilli()
		msg.Return(true)
	default:
		msg.Return(false)
	}
	return &multiplex.HookState{Handled: true}
}

// NotifyParams contain parameters of all commands of Notify.
type NotifyParams struct {
	To      string
	Message string
}

// Extract parameters from a service message and return new NotifyParams.
func NewNotifyParams(msg *multiplex.ServiceMessage) *NotifyParams {
	return &NotifyParams{
		To:      msg.GetParam("to", "").(string),
		Message: msg.GetParam("message", "").(string),
	}
}
