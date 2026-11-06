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
	"fmt"
)

// TraceTxError describes one transaction of a debug_traceBlockByNumber result replaced with an error entry.
type TraceTxError struct {
	Index  int
	TxHash string
	Detail string
}

// Check eth_getBlockByNumber result for an empty attestor field: key missing, null, empty string, or bare "0x".
func DetectEmptyAttestor(blockRaw json.RawMessage) (found bool, value string, err error) {
	var block struct {
		Attestor json.RawMessage `json:"attestor"`
	}
	if unmarshalErr := json.Unmarshal(blockRaw, &block); unmarshalErr != nil {
		return false, "", fmt.Errorf("unmarshal block for attestor check: %w", unmarshalErr)
	}
	if len(block.Attestor) == 0 {
		return true, "missing", nil
	}
	if string(block.Attestor) == "null" {
		return true, "null", nil
	}
	var attestor string
	if unmarshalErr := json.Unmarshal(block.Attestor, &attestor); unmarshalErr != nil {
		return true, fmt.Sprintf("non-string %s", string(block.Attestor)), nil
	}
	if attestor == "" || attestor == "0x" {
		return true, fmt.Sprintf("%q", attestor), nil
	}
	return false, "", nil
}

// Checkdebug_traceBlockByNumber result for transactions the node reported as error entries instead of traces.
func DetectTraceTxErrors(traceRaw json.RawMessage) ([]TraceTxError, error) {
	var traces []struct {
		TxHash string  `json:"txHash"`
		Error  *string `json:"error"`
	}
	if unmarshalErr := json.Unmarshal(traceRaw, &traces); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal traces for tx error check: %w", unmarshalErr)
	}
	txErrors := []TraceTxError{}
	for i, trace := range traces {
		if trace.Error == nil {
			continue
		}
		txErrors = append(txErrors, TraceTxError{
			Index:  i,
			TxHash: trace.TxHash,
			Detail: *trace.Error,
		})
	}
	return txErrors, nil
}
