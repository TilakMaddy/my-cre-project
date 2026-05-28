package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func RPCCall(url, method string, params []any) (json.RawMessage, error) {
	batch := []rpcRequest{
		{JSONRPC: "2.0", Method: method, Params: params, ID: 1},
	}
	body, err := json.Marshal(batch)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out []rpcResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode rpc: %w", err)
	}
	if len(out) == 0 {
		return nil, errors.New("empty rpc batch")
	}
	if out[0].Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", out[0].Error.Code, out[0].Error.Message)
	}
	return out[0].Result, nil
}
