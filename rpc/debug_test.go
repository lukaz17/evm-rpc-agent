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
	"testing"
)

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
