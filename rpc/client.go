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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const jsonRPCVersion = "2.0"

// Request represents a JSON-RPC request.
type Request struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
	ID      int           `json:"id"`
}

// Response represents a JSON-RPC response.
type Response[T any] struct {
	JSONRPC   string `json:"jsonrpc"`
	Result    *T     `json:"result"`
	RawResult string `json:"-"`
	Error     *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID int `json:"id"`
}

// Client represents a JSON-RPC client.
type Client struct {
	endpoint  string
	client    *http.Client
	requestID int
}

// Return new instance of Client from an endpoint URL.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint:  endpoint,
		client:    &http.Client{},
		requestID: 0,
	}
}

// Call performs a JSON-RPC method call.
func (c *Client) Call(result any, method string, params ...any) error {
	c.requestID++

	req := Request{
		JSONRPC: jsonRPCVersion,
		Method:  method,
		Params:  params,
		ID:      c.requestID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.client.Post(c.endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	var rpcResp Response[json.RawMessage]
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if result == nil || rpcResp.Result == nil {
		return nil
	}

	rpcResp.RawResult = string(*rpcResp.Result)
	if err := json.Unmarshal(*rpcResp.Result, result); err != nil {
		return fmt.Errorf("unmarshal result: %w", err)
	}

	return nil
}
