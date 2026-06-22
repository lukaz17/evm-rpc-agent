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

package rpc

import (
	"testing"
)

func TestTraceBlockByNumber_DefaultTracer(t *testing.T) {
	c := newTestClient(t)
	traces, err := TraceBlockByNumber(c, "0x1a33b7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, tr := range traces {
		if tr.Type != TraceTypeDefault {
			t.Fatalf("traces[%d] type = %q, want %q", i, tr.Type, TraceTypeDefault)
		}
	}
	t.Logf("block 0x1a33b7 trace count: %d", len(traces))
}

func TestTraceBlockByNumber_WithCallTracer(t *testing.T) {
	c := newTestClient(t)
	tracer := "callTracer"
	opts := TraceBlockOptions{Tracer: &tracer}
	traces, err := TraceBlockByNumber(c, "latest", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, tr := range traces {
		if tr.Type != TraceTypeCall {
			t.Fatalf("traces[%d] type = %q, want %q", i, tr.Type, TraceTypeCall)
		}
	}
	t.Logf("callTracer trace count: %d", len(traces))
}

func TestTraceBlockByNumber_WithPrestateTracer(t *testing.T) {
	c := newTestClient(t)
	tracer := "prestateTracer"
	opts := TraceBlockOptions{Tracer: &tracer}
	traces, err := TraceBlockByNumber(c, "latest", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, tr := range traces {
		if tr.Type != TraceTypePrestate {
			t.Fatalf("traces[%d] type = %q, want %q", i, tr.Type, TraceTypePrestate)
		}
	}
	t.Logf("prestateTracer trace count: %d", len(traces))
}

func TestTraceBlockByNumber_NonExistent(t *testing.T) {
	c := newTestClient(t)
	traces, err := TraceBlockByNumber(c, "0xffffffff")
	if traces != nil {
		t.Fatal("expected nil traces for non-existent block")
	}
	if err == nil {
		t.Log("node returned nil without error (acceptable)")
		return
	}
	t.Logf("got expected error: %v", err)
}

func TestTraceBlockByNumber_WithCustomTracerDefault(t *testing.T) {
	c := newTestClient(t)
	unknown := "unknownTracer"
	opts := TraceBlockOptions{Tracer: &unknown}
	traces, err := TraceBlockByNumber(c, "0x1a33b7", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, tr := range traces {
		if tr.Type != TraceTypeDefault {
			t.Fatalf("traces[%d] type = %q, want %q", i, tr.Type, TraceTypeDefault)
		}
	}
	t.Logf("unknown tracer trace count: %d", len(traces))
}

func TestTraceTypeFromOpts(t *testing.T) {
	tests := []struct {
		name string
		opts []TraceBlockOptions
		want string
	}{
		{
			name: "no options",
			opts: nil,
			want: TraceTypeDefault,
		},
		{
			name: "empty options",
			opts: []TraceBlockOptions{},
			want: TraceTypeDefault,
		},
		{
			name: "no tracer set",
			opts: []TraceBlockOptions{{}},
			want: TraceTypeDefault,
		},
		{
			name: "callTracer",
			opts: func() []TraceBlockOptions {
				t := "callTracer"
				return []TraceBlockOptions{{Tracer: &t}}
			}(),
			want: TraceTypeCall,
		},
		{
			name: "prestateTracer",
			opts: func() []TraceBlockOptions {
				t := "prestateTracer"
				return []TraceBlockOptions{{Tracer: &t}}
			}(),
			want: TraceTypePrestate,
		},
		{
			name: "unknown tracer defaults to default",
			opts: func() []TraceBlockOptions {
				t := "someRandomTracer"
				return []TraceBlockOptions{{Tracer: &t}}
			}(),
			want: TraceTypeDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := traceTypeFromOpts(tt.opts)
			if got != tt.want {
				t.Fatalf("traceTypeFromOpts() = %q, want %q", got, tt.want)
			}
		})
	}
}
