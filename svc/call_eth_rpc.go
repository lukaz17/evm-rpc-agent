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
	"encoding/json"
	"strings"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
	"github.com/tforce-io/tf-golib/opx"
)

// CallEthRpc executes low-level Ethereum RPC call.
type CallEthRpc struct {
	multiplex.ServiceCore
	i   *multiplex.ServiceCoreInternal
	o   *NetworkOptions
	rpc *rpc.Client
}

// Return new CallEthRpc instance.
func NewCallEthRpc(name string, rpc *rpc.Client, cfg *config.ServiceConfig, logger zerolog.Logger) *CallEthRpc {
	svc := &CallEthRpc{
		rpc: rpc,
	}
	svc.i = svc.InitServiceCore(name, config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	svc.o = &NetworkOptions{
		MaxRetries:  cfg.MaxRpcRrtryCount,
		MaxRetryGap: 200 * 1000000,
	}
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *CallEthRpc) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "rpc_call":
		p := NewCallEthRpcParams(msg)
		var raw json.RawMessage
		err := s.rpc.Call(&raw, p.Method, p.Params...)
		retryCount := uint64(0)
		halfRetry := false
		for err != nil && retryCount < s.o.MaxRetries {
			errStr := err.Error()
			if strings.HasPrefix(errStr, "503 Service Unavailable: <html><body><h1>503 Service Unavailable</h1>") {
				if !halfRetry {
					retryCount--
				}
				halfRetry = !halfRetry
			} else {
				s.i.Logger.Warnf("%s#%02d: %s retrying. %v", s.i.ServiceID, workerID, p.Method, errStr)
			}
			s.o.WaitRetryGap()
			err = s.rpc.Call(&raw, p.Method, p.Params...)
			retryCount++
		}
		result := &CallEthRpcResult{
			Method:     p.Method,
			Params:     p.Params,
			Data:       raw,
			Error:      err,
			RetryCount: retryCount,
		}
		s.i.Logger.Infof("%s#%02d: %s processed. %s. Retry count = %d.", s.i.ServiceID, workerID, p.Method,
			opx.Ternary(err == nil, "SUCCESS", "FAILED"),
			retryCount,
		)
		msg.Return(result)
	default:
		s.i.Logger.Warnf("%s#%02d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

// CallEthRpcParams contain parameters of all commands of CallEthRpc.
type CallEthRpcParams struct {
	Method string
	Params []any
}

// Extract parameters from a service message and return new CallEthRpcParams.
func NewCallEthRpcParams(msg *multiplex.ServiceMessage) *CallEthRpcParams {
	return &CallEthRpcParams{
		Method: msg.GetParam("method", "").(string),
		Params: msg.GetParam("params", []any{}).([]any),
	}
}

// CallEthRpcResult contains the results of all commands of CallEthRpc.
type CallEthRpcResult struct {
	Method     string
	Params     []any
	Data       json.RawMessage
	Error      error
	RetryCount uint64
}
