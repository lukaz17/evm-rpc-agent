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

package svc

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendAnomaly(t *testing.T) {
	recordedAt := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)

	t.Run("writes header then row for a new file", func(t *testing.T) {
		dir := t.TempDir()
		p := &RecordAnomalyParams{
			Type:        AnomalyTypeTraceTxError,
			BlockNumber: "1717172",
			TxIndex:     3,
			TxHash:      "0xabc",
			Detail:      "execution timeout",
		}
		if err := appendAnomaly(dir, recordedAt, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := anomalyTableHeader + "| 2026-11-05T10:00:00Z | 1717172 | 3 | 0xabc | execution timeout |\n"
		assertFileContent(t, filepath.Join(dir, AnomalyTypeTraceTxError+".md"), want)
	})

	t.Run("appends without duplicating the header", func(t *testing.T) {
		dir := t.TempDir()
		p := &RecordAnomalyParams{Type: AnomalyTypeTraceTxError, BlockNumber: "1", TxIndex: 0, Detail: "execution timeout"}
		if err := appendAnomaly(dir, recordedAt, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := appendAnomaly(dir, recordedAt, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content := readFile(t, filepath.Join(dir, AnomalyTypeTraceTxError+".md"))
		if got := strings.Count(content, "| time |"); got != 1 {
			t.Errorf("header count = %d, want 1", got)
		}
		if got := strings.Count(content, "execution timeout"); got != 2 {
			t.Errorf("row count = %d, want 2", got)
		}
	})

	t.Run("separates anomaly types into files", func(t *testing.T) {
		dir := t.TempDir()
		if err := appendAnomaly(dir, recordedAt, &RecordAnomalyParams{Type: AnomalyTypeEmptyAttestor, BlockNumber: "7", TxIndex: -1, Detail: "attestor field is null"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := appendAnomaly(dir, recordedAt, &RecordAnomalyParams{Type: AnomalyTypeTraceTxError, BlockNumber: "8", TxIndex: 0, Detail: "execution timeout"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertFileContent(t, filepath.Join(dir, AnomalyTypeEmptyAttestor+".md"),
			anomalyTableHeader+"| 2026-11-05T10:00:00Z | 7 |  |  | attestor field is null |\n")
		assertFileContent(t, filepath.Join(dir, AnomalyTypeTraceTxError+".md"),
			anomalyTableHeader+"| 2026-11-05T10:00:00Z | 8 | 0 |  | execution timeout |\n")
	})

	t.Run("empty dir writes no file", func(t *testing.T) {
		p := &RecordAnomalyParams{Type: AnomalyTypeEmptyAttestor, BlockNumber: "7", TxIndex: -1, Detail: "attestor field is null"}
		if err := appendAnomaly("", recordedAt, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects type that is not a plain name", func(t *testing.T) {
		dir := t.TempDir()
		p := &RecordAnomalyParams{Type: "../escape", BlockNumber: "7", TxIndex: -1, Detail: "d"}
		if err := appendAnomaly(dir, recordedAt, p); err == nil {
			t.Fatal("expected error, got nil")
		}
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatalf("read dir: %v", readErr)
		}
		if len(entries) != 0 {
			t.Errorf("file count = %d, want 0", len(entries))
		}
	})

	t.Run("rejects invalid block number", func(t *testing.T) {
		dir := t.TempDir()
		p := &RecordAnomalyParams{Type: AnomalyTypeEmptyAttestor, BlockNumber: "seven", TxIndex: -1, Detail: "d"}
		if err := appendAnomaly(dir, recordedAt, p); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("escapes cell content", func(t *testing.T) {
		dir := t.TempDir()
		p := &RecordAnomalyParams{Type: AnomalyTypeTraceTxError, BlockNumber: "9", TxIndex: 0, Detail: "a|b\nc"}
		if err := appendAnomaly(dir, recordedAt, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content := readFile(t, filepath.Join(dir, AnomalyTypeTraceTxError+".md"))
		if !strings.Contains(content, `a\|b c`) {
			t.Errorf("content %q misses escaped cell", content)
		}
		if lineCount := strings.Count(content, "\n"); lineCount != 3 {
			t.Errorf("line count = %d, want 3", lineCount)
		}
	})
}

func TestEmptyAttestorRecords(t *testing.T) {
	blockNumber := big.NewInt(1717171)

	tests := []struct {
		name    string
		block   string
		records int
		detail  string
	}{
		{
			name:    "valid attestor yields no record",
			block:   `{"number":"0x1a33b3","attestor":"0x1234567890abcdef1234567890abcdef12345678"}`,
			records: 0,
		},
		{
			name:    "missing attestor key yields one record",
			block:   `{"number":"0x1a33b3"}`,
			records: 1,
			detail:  "attestor field is missing",
		},
		{
			name:    "empty attestor yields one record",
			block:   `{"number":"0x1a33b3","attestor":""}`,
			records: 1,
			detail:  `attestor field is ""`,
		},
		{
			name:    "null block result yields no record",
			block:   `null`,
			records: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := emptyAttestorRecords(blockNumber, json.RawMessage(tt.block))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(records) != tt.records {
				t.Fatalf("record count = %d, want %d", len(records), tt.records)
			}
			if tt.records == 0 {
				return
			}
			record := records[0]
			if record.Type != AnomalyTypeEmptyAttestor {
				t.Errorf("type = %q, want %q", record.Type, AnomalyTypeEmptyAttestor)
			}
			if record.BlockNumber != "1717171" {
				t.Errorf("block number = %q, want %q", record.BlockNumber, "1717171")
			}
			if record.TxIndex != -1 {
				t.Errorf("tx index = %d, want -1", record.TxIndex)
			}
			if record.Detail != tt.detail {
				t.Errorf("detail = %q, want %q", record.Detail, tt.detail)
			}
		})
	}
}

func TestTraceTxErrorRecords(t *testing.T) {
	blockNumber := big.NewInt(1717172)

	tests := []struct {
		name    string
		traces  string
		records []RecordAnomalyParams
	}{
		{
			name: "traced transactions yield no records",
			traces: `[
				{"txHash":"0xaaa","result":{"type":"CALL"}}
			]`,
			records: []RecordAnomalyParams{},
		},
		{
			name: "timed out transaction yields one record",
			traces: `[
				{"txHash":"0xaaa","result":{"type":"CALL"}},
				{"error":"execution timeout"}
			]`,
			records: []RecordAnomalyParams{
				{Type: AnomalyTypeTraceTxError, BlockNumber: "1717172", TxIndex: 1, Detail: "execution timeout"},
			},
		},
		{
			name:    "null trace result yields no records",
			traces:  `null`,
			records: []RecordAnomalyParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := traceTxErrorRecords(blockNumber, json.RawMessage(tt.traces))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(records) != len(tt.records) {
				t.Fatalf("record count = %d, want %d", len(records), len(tt.records))
			}
			for i, record := range records {
				want := tt.records[i]
				if *record != want {
					t.Errorf("records[%d] = %+v, want %+v", i, *record, want)
				}
			}
		})
	}
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()
	if got := readFile(t, path); got != want {
		t.Errorf("content of %s = %q, want %q", path, got, want)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
