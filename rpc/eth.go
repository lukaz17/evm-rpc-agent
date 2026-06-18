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
	"strconv"
)

// Return the latest block number.
func BlockNumber(c *Client) (uint64, error) {
	var result string
	if err := c.Call(&result, "eth_blockNumber"); err != nil {
		return 0, fmt.Errorf("eth_blockNumber: %w", err)
	}

	num, err := strconv.ParseUint(result[2:], 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parse block number %s: %w", result, err)
	}

	return num, nil
}

// Return block by block number.
func GetBlockByNumber(c *Client, blockNum string, fullTx bool) (*Block, error) {
	var raw json.RawMessage
	if err := c.Call(&raw, "eth_getBlockByNumber", blockNum, fullTx); err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber: %w", err)
	}

	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, fmt.Errorf("unmarshal block: %w", err)
	}

	return &block, nil
}
