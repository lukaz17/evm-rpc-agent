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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadJson(t *testing.T, name string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "rpc", "blobs", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal %s: %v", name, err)
	}
	return raw
}

func loadJsonN(t *testing.T, name string) map[string][]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "rpc", "blobs", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	var raw map[string][]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal %s: %v", name, err)
	}
	return raw
}
