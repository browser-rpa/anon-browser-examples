// Command plugin demonstrating the interactive anon-browser capability
// protocol. The executable itself is written in Go; Core remains responsible
// for browser, variable and network operations.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type invocation struct {
	NodeType string         `json:"node_type"`
	Input    map[string]any `json:"input"`
}

type capabilityCall struct {
	Type       string         `json:"type"`
	ID         string         `json:"id"`
	Capability string         `json:"capability"`
	Action     string         `json:"action"`
	Arguments  map[string]any `json:"arguments,omitempty"`
}

type capabilityResult struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type nodeResult struct {
	OK     bool `json:"ok"`
	Result *struct {
		Output    map[string]any `json:"output,omitempty"`
		Variables map[string]any `json:"variables,omitempty"`
		NextPort  string         `json:"next_port"`
	} `json:"result,omitempty"`
	Error string `json:"error,omitempty"`
}

func main() {
	decoder := json.NewDecoder(bufio.NewReader(os.Stdin))
	encoder := json.NewEncoder(os.Stdout)
	var request invocation
	if err := decoder.Decode(&request); err != nil {
		writeError(encoder, fmt.Errorf("decode invocation: %w", err))
		return
	}

	var result any
	var err error
	switch request.NodeType {
	case "examples.capability.page_title":
		result, err = call(encoder, decoder, "browser", "read_title", nil)
	case "examples.capability.page_url":
		result, err = call(encoder, decoder, "browser", "read_url", nil)
	case "examples.capability.variable_get":
		result, err = call(encoder, decoder, "variables", "get", map[string]any{"name": request.Input["name"]})
	case "examples.capability.variable_set":
		result, err = call(encoder, decoder, "variables", "set", map[string]any{
			"name":  request.Input["name"],
			"value": request.Input["value"],
		})
	case "examples.capability.http_request":
		result, err = call(encoder, decoder, "http", "request", map[string]any{
			"method":  request.Input["method"],
			"url":     request.Input["url"],
			"headers": request.Input["headers"],
			"body":    request.Input["body"],
		})
	default:
		err = fmt.Errorf("unsupported node type %q", request.NodeType)
	}
	if err != nil {
		writeError(encoder, err)
		return
	}

	output := map[string]any{"value": result}
	variables := map[string]any{}
	if name, ok := request.Input["output_variable"].(string); ok && name != "" {
		variables[name] = result
	}
	writeResult(encoder, output, variables)
}

func call(encoder *json.Encoder, decoder *json.Decoder, capability, action string, arguments map[string]any) (any, error) {
	if err := encoder.Encode(capabilityCall{Type: "capability_call", ID: "call-1", Capability: capability, Action: action, Arguments: arguments}); err != nil {
		return nil, fmt.Errorf("send capability call: %w", err)
	}
	var response capabilityResult
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("decode capability result: %w", err)
	}
	if response.Type != "capability_result" || response.ID != "call-1" {
		return nil, errors.New("invalid capability result")
	}
	if !response.OK {
		if response.Error == "" {
			response.Error = "capability call failed"
		}
		return nil, errors.New(response.Error)
	}
	return response.Data, nil
}

func writeResult(encoder *json.Encoder, output, variables map[string]any) {
	_ = encoder.Encode(nodeResult{OK: true, Result: &struct {
		Output    map[string]any `json:"output,omitempty"`
		Variables map[string]any `json:"variables,omitempty"`
		NextPort  string         `json:"next_port"`
	}{Output: output, Variables: variables, NextPort: "success"}})
}

func writeError(encoder *json.Encoder, err error) {
	if err == nil {
		err = io.ErrUnexpectedEOF
	}
	_ = encoder.Encode(nodeResult{OK: false, Error: err.Error()})
}
