// Echo demonstrates shared protocol types, Go node descriptors and SDK serving.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	sdk "github.com/browser-rpa/anon-browser-sdk/plugin"
)

const nodeType = "examples.echo"

func manifest() (sdk.Manifest, error) {
	node, err := sdk.Describe(sdk.NodeMetadata{Type: nodeType, Version: 1, DisplayName: "Echo 输入", Description: "把输入文本输出到当前 Run", Category: "示例", SecurityLevel: "sensitive", Idempotency: "idempotent"},
		sdk.Object(map[string]sdk.Schema{"value": sdk.String("输入值")}, "value"),
		sdk.Object(map[string]sdk.Schema{"value": sdk.String("结果")}, "value"),
		sdk.NodePolicy{RequiredPermissions: []string{sdk.PermissionVariablesWrite}, Capabilities: []string{sdk.PermissionVariablesWrite}, Timeouts: sdk.NodeTimeout{Default: 10 * time.Second, Minimum: time.Millisecond * 100, Maximum: time.Minute}, MaxAttempts: 1})
	return sdk.Manifest{ID: nodeType, Name: "Echo 示例节点", Version: "0.2.0", Kind: sdk.KindAutomation, Entrypoint: "bin/plugin", Permissions: []string{sdk.PermissionVariablesWrite}, Nodes: []sdk.NodeDescriptor{node}}, err
}
func handle(_ context.Context, call sdk.NodeCall) (sdk.NodeResult, error) {
	value, ok := call.Invocation.Input["value"].(string)
	if !ok {
		return sdk.NodeResult{}, fmt.Errorf("value must be a string")
	}
	return sdk.NodeResult{NextPort: "success", Output: map[string]any{"value": value}, Variables: map[string]any{"echo_value": value}}, nil
}
func run() error {
	m, err := manifest()
	if err != nil {
		return err
	}
	if err := m.Validate(); err != nil {
		return err
	}
	if len(os.Args) == 2 && os.Args[1] == "--manifest" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(m)
	}
	server := sdk.NewServer()
	if err := server.Register(nodeType, 1, handle); err != nil {
		return err
	}
	return server.Serve(context.Background(), os.Stdin, os.Stdout)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
