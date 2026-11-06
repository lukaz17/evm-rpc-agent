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
	"testing"
)

func TestTraceBlockTracerOptions(t *testing.T) {
	tests := []struct {
		name           string
		timeoutSeconds uint64
		wantTracer     string
		wantTimeout    string
	}{
		{
			name:           "zero timeout omits the param",
			timeoutSeconds: 0,
			wantTracer:     "callTracer",
			wantTimeout:    "",
		},
		{
			name:           "default 120 seconds",
			timeoutSeconds: 120,
			wantTracer:     "callTracer",
			wantTimeout:    "120s",
		},
		{
			name:           "custom 90 seconds",
			timeoutSeconds: 90,
			wantTracer:     "callTracer",
			wantTimeout:    "90s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := traceBlockTracerOptions(tt.timeoutSeconds)

			if opts.Tracer == nil || *opts.Tracer != tt.wantTracer {
				t.Errorf("tracer = %v, want %q", opts.Tracer, tt.wantTracer)
			}

			if tt.wantTimeout == "" {
				if opts.Timeout != nil {
					t.Errorf("timeout = %q, want omitted", *opts.Timeout)
				}
			} else if opts.Timeout == nil || *opts.Timeout != tt.wantTimeout {
				t.Errorf("timeout = %v, want %q", opts.Timeout, tt.wantTimeout)
			}
		})
	}
}

// Pin the JSON params of the debug_traceBlockByNumber call so the wire format matches what the node accepts.
func TestTraceBlockTracerOptions_JsonParams(t *testing.T) {
	encoded, err := json.Marshal(traceBlockTracerOptions(120))
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	want := `{"tracer":"callTracer","timeout":"120s"}`
	if string(encoded) != want {
		t.Errorf("encoded params = %s, want %s", encoded, want)
	}

	encodedZero, err := json.Marshal(traceBlockTracerOptions(0))
	if err != nil {
		t.Fatalf("marshal zero options: %v", err)
	}

	wantZero := `{"tracer":"callTracer"}`
	if string(encodedZero) != wantZero {
		t.Errorf("encoded zero params = %s, want %s", encodedZero, wantZero)
	}
}
