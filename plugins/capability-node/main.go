// All wire types, response correlation and IO limits come from the public SDK.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	sdk "github.com/browser-rpa/anon-browser-sdk/plugin"
)

const prefix = "examples.capability."

var actions = []string{"page_title", "page_url", "variable_get", "variable_set", "http_request"}

func manifest() (sdk.Manifest, error) {
	manifest := sdk.Manifest{ID: "examples.capability", Name: "Capability 调用示例", Version: "0.2.0", Kind: sdk.KindAutomation, Entrypoint: "bin/plugin", Permissions: []string{sdk.PermissionBrowserRead, sdk.PermissionVariablesRead, sdk.PermissionVariablesWrite, sdk.PermissionNetworkFetch}}
	for _, action := range actions {
		properties := map[string]sdk.Schema{"output_variable": sdk.String("保存到变量")}
		var required []string
		permission := sdk.PermissionBrowserRead
		security, idempotency := "sensitive", "idempotent"
		resources := sdk.NodeResources{}
		switch action {
		case "page_title", "page_url":
			resources = sdk.NodeResources{ProfileLease: true, Runtime: true, ActivePage: true}
		case "variable_get", "variable_set":
			properties["name"] = sdk.String("变量名")
			required = []string{"name"}
			permission = sdk.PermissionVariablesRead
			if action == "variable_set" {
				properties["value"] = sdk.Schema{Title: "变量值"}
				required = append(required, "value")
				permission = sdk.PermissionVariablesWrite
			}
		case "http_request":
			properties["method"] = sdk.Schema{Type: "string", Title: "请求方法", Enum: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}}
			properties["url"] = sdk.String("接口地址")
			properties["headers"] = sdk.Schema{Type: "object"}
			properties["body"] = sdk.Schema{}
			required = []string{"method", "url"}
			permission = sdk.PermissionNetworkFetch
			security = "high_risk"
			idempotency = "non_idempotent"
		}
		permissions := []string{permission}
		// Every node supports output_variable, so returning variables requires
		// a write grant even when its primary operation only reads data.
		if permission != sdk.PermissionVariablesWrite {
			permissions = append(permissions, sdk.PermissionVariablesWrite)
		}
		node, err := sdk.Describe(sdk.NodeMetadata{Type: prefix + action, Version: 1, DisplayName: action + "（插件）", Category: "插件示例", Description: "通过宿主授权能力调用，不直连浏览器", SecurityLevel: security, Idempotency: idempotency},
			sdk.Object(properties, required...), sdk.Object(map[string]sdk.Schema{"value": {}}),
			sdk.NodePolicy{RequiredPermissions: permissions, Capabilities: permissions, Resources: resources, Timeouts: sdk.NodeTimeout{Default: 10 * time.Second, Minimum: 100 * time.Millisecond, Maximum: time.Minute}, MaxAttempts: 1})
		if err != nil {
			return sdk.Manifest{}, err
		}
		manifest.Nodes = append(manifest.Nodes, node)
	}
	return manifest, nil
}
func handle(ctx context.Context, call sdk.NodeCall) (sdk.NodeResult, error) {
	var value any
	var err error
	switch call.Invocation.NodeType {
	case prefix + "page_title":
		value, err = call.Client.ReadTitle(ctx)
	case prefix + "page_url":
		value, err = call.Client.ReadURL(ctx)
	case prefix + "variable_get":
		value, err = call.Client.GetVariable(ctx, call.String("name"))
	case prefix + "variable_set":
		err = call.Client.SetVariable(ctx, call.String("name"), call.Invocation.Input["value"])
	case prefix + "http_request":
		var request sdk.HTTPRequest
		if err = call.DecodeInput(&request); err == nil {
			value, err = call.Client.HTTPRequest(ctx, request)
		}
	default:
		return sdk.NodeResult{}, fmt.Errorf("unsupported node %q", call.Invocation.NodeType)
	}
	if err != nil {
		return sdk.NodeResult{}, err
	}
	result := sdk.NodeResult{NextPort: "success", Output: map[string]any{"value": value}}
	if output := call.String("output_variable"); output != "" {
		result.Variables = map[string]any{output: value}
	}
	return result, nil
}
func run() error {
	manifest, err := manifest()
	if err != nil {
		return err
	}
	if err := manifest.Validate(); err != nil {
		return err
	}
	if len(os.Args) == 2 && os.Args[1] == "--manifest" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(manifest)
	}
	server := sdk.NewServer()
	for _, node := range manifest.Nodes {
		if err = server.Register(node.Metadata.Type, node.Metadata.Version, handle); err != nil {
			return err
		}
	}
	return server.Serve(context.Background(), os.Stdin, os.Stdout)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
