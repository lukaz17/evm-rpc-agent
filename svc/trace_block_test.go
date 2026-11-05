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
	"math/big"
	"testing"

	"github.com/tforce-io/tf-golib/multiplex"
)

func TestTraceBlock_ParamsExtraction(t *testing.T) {
	tests := []struct {
		name      string
		params    multiplex.ExecParams
		wantFrom  int64
		wantTo    int64
		wantBatch uint64
	}{
		{
			name:      "default values",
			params:    multiplex.ExecParams{},
			wantFrom:  0,
			wantTo:    0,
			wantBatch: 1,
		},
		{
			name: "explicit values",
			params: multiplex.ExecParams{
				"from_block_number": big.NewInt(1000),
				"to_block_number":   big.NewInt(2000),
				"batch_size":        uint64(50),
			},
			wantFrom:  1000,
			wantTo:    2000,
			wantBatch: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &multiplex.ServiceMessage{
				Params: tt.params,
			}
			p := NewTraceBlockParams(msg)

			if p.FromBlockNumber.Int64() != tt.wantFrom {
				t.Errorf("from_block_number = %d, want %d", p.FromBlockNumber.Int64(), tt.wantFrom)
			}
			if p.ToBlockNumber.Int64() != tt.wantTo {
				t.Errorf("to_block_number = %d, want %d", p.ToBlockNumber.Int64(), tt.wantTo)
			}
			if p.BatchSize != tt.wantBatch {
				t.Errorf("batch_size = %d, want %d", p.BatchSize, tt.wantBatch)
			}
		})
	}
}
