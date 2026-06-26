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

package core

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// Decode a hex string with 0x prefix into bytes.
func decodeHex(s string) ([]byte, error) {
	if s == "" || s == "0x" {
		return nil, nil
	}
	src := strings.TrimPrefix(s, "0x")
	if len(src)%2 != 0 {
		src = "0" + src
	}
	b, err := hex.DecodeString(src)
	if err != nil {
		return nil, fmt.Errorf("decode hex %s: %w", s, err)
	}
	return b, nil
}

// Decode a numeric string into big.Int.
func DecodeNumericString(s string) (*big.Int, error) {
	if s == "" || s == "0x" {
		return nil, nil
	}
	src := strings.TrimPrefix(s, "0x")
	base := 16
	if src == s {
		base = 10
	}
	val, ok := new(big.Int).SetString(src, base)
	if !ok {
		return nil, fmt.Errorf("decode numeric string %s: invalid number", s)
	}
	return val, nil
}
