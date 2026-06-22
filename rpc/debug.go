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
	"encoding/json"
	"fmt"
)

// TraceBlockOptions specifies optional parameters for debug trace block calls.
type TraceBlockOptions struct {
	Tracer       *string         `json:"tracer,omitempty"`
	Timeout      *string         `json:"timeout,omitempty"`
	TracerConfig json.RawMessage `json:"tracerConfig,omitempty"`
}

// TraceBlockByNumber calls debug_traceBlockByNumber and returns per-transaction trace results.
func TraceBlockByNumber(c *Client, blockNum string, opts ...TraceBlockOptions) ([]TransactionTrace, error) {
	var raw json.RawMessage

	params := []any{blockNum}
	if len(opts) > 0 {
		params = append(params, opts[0])
	}

	if err := c.Call(&raw, "debug_traceBlockByNumber", params...); err != nil {
		return nil, fmt.Errorf("debug_traceBlockByNumber: %w", err)
	}

	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var traces []TransactionTrace
	if err := json.Unmarshal(raw, &traces); err != nil {
		return nil, fmt.Errorf("unmarshal traces: %w", err)
	}

	traceType := traceTypeFromOpts(opts)
	for i := range traces {
		traces[i].Type = traceType
	}

	return traces, nil
}

func traceTypeFromOpts(opts []TraceBlockOptions) string {
	if len(opts) == 0 || opts[0].Tracer == nil {
		return TraceTypeDefault
	}
	switch *opts[0].Tracer {
	case "callTracer":
		return TraceTypeCall
	case "prestateTracer":
		return TraceTypePrestate
	default:
		return TraceTypeDefault
	}
}
