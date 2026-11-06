// Copyright (C) 2025  Nguyen Nhat Tung
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package rpc

import (
	"encoding/json"
	"testing"
)

func TestDetectEmptyAttestor(t *testing.T) {
	tests := []struct {
		name  string
		block string
		found bool
		value string
		err   bool
	}{
		{
			name:  "valid attestor",
			block: `{"number":"0x1","attestor":"0x1234567890abcdef1234567890abcdef12345678"}`,
			found: false,
		},
		{
			name:  "missing attestor key",
			block: `{"number":"0x1"}`,
			found: true,
			value: "missing",
		},
		{
			name:  "null attestor",
			block: `{"number":"0x1","attestor":null}`,
			found: true,
			value: "null",
		},
		{
			name:  "empty string attestor",
			block: `{"number":"0x1","attestor":""}`,
			found: true,
			value: `""`,
		},
		{
			name:  "bare hex prefix attestor",
			block: `{"number":"0x1","attestor":"0x"}`,
			found: true,
			value: `"0x"`,
		},
		{
			name:  "non-string attestor",
			block: `{"number":"0x1","attestor":123}`,
			found: true,
			value: "non-string 123",
		},
		{
			name:  "malformed block json",
			block: `{"number":`,
			found: false,
			err:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, value, err := DetectEmptyAttestor(json.RawMessage(tt.block))
			if tt.err {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
			if value != tt.value {
				t.Errorf("value = %q, want %q", value, tt.value)
			}
		})
	}
}

func TestDetectTraceTxErrors(t *testing.T) {
	tests := []struct {
		name     string
		traces   string
		txErrors []TraceTxError
		err      bool
	}{
		{
			name: "all transactions traced",
			traces: `[
				{"txHash":"0xaaa","result":{"type":"CALL"}},
				{"txHash":"0xbbb","result":{"type":"CALL"}}
			]`,
			txErrors: []TraceTxError{},
		},
		{
			name: "one timed out transaction without tx hash",
			traces: `[
				{"txHash":"0xaaa","result":{"type":"CALL"}},
				{"error":"execution timeout"}
			]`,
			txErrors: []TraceTxError{
				{Index: 1, Detail: "execution timeout"},
			},
		},
		{
			name: "error transaction with tx hash",
			traces: `[
				{"txHash":"0xaaa","error":"execution timeout"}
			]`,
			txErrors: []TraceTxError{
				{Index: 0, TxHash: "0xaaa", Detail: "execution timeout"},
			},
		},
		{
			name: "multiple error transactions mixed with traces",
			traces: `[
				{"txHash":"0xaaa","result":{"type":"CALL"}},
				{"error":"execution timeout"},
				{"txHash":"0xccc","result":{"type":"CALL"}},
				{"error":"tracer exceeded timeout"}
			]`,
			txErrors: []TraceTxError{
				{Index: 1, Detail: "execution timeout"},
				{Index: 3, Detail: "tracer exceeded timeout"},
			},
		},
		{
			name:     "empty block result",
			traces:   `[]`,
			txErrors: []TraceTxError{},
		},
		{
			name:     "null result",
			traces:   `null`,
			txErrors: []TraceTxError{},
		},
		{
			name:   "non-array result",
			traces: `{"error":"execution timeout"}`,
			err:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txErrors, err := DetectTraceTxErrors(json.RawMessage(tt.traces))
			if tt.err {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(txErrors) != len(tt.txErrors) {
				t.Fatalf("tx error count = %d, want %d", len(txErrors), len(tt.txErrors))
			}
			for i, txError := range txErrors {
				if txError != tt.txErrors[i] {
					t.Errorf("txErrors[%d] = %+v, want %+v", i, txError, tt.txErrors[i])
				}
			}
		})
	}
}
