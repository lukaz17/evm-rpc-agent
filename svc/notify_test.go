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

	"github.com/lukaz17/evm-rpc-agent/mem"
	"github.com/tforce-io/tf-golib/multiplex"
)

func TestNotify_LastUpdate(t *testing.T) {
	mem.LastUpdate = 0

	ctrl := newTestController(t, nil)
	msg := multiplex.ExecParams{}
	ctrl.DispatchOnce("Notify", "last_update", msg)

	result := msg.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}
	if mem.LastUpdate == 0 {
		t.Fatal("expected LastUpdate to be updated, got 0")
	}
}

func TestNotify_UnknownCommand(t *testing.T) {
	ctrl := newTestController(t, nil)
	msg := multiplex.ExecParams{}
	ctrl.DispatchOnce("Notify", "unknown_cmd", msg)

	result := msg.ReturnResult()
	if result != false {
		t.Fatalf("expected return value false, got %v", result)
	}
}
