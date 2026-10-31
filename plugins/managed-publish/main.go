// managed-publish demonstrates a complete publishing operation using only
// host-authorized capabilities. The companion fixture uses no account data.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	sdk "github.com/browser-rpa/anon-browser-sdk/plugin"
)

const nodeType = "examples.managed_publish"

type publishInput struct {
	Origin string            `json:"origin"`
	Title  string            `json:"title"`
	Body   string            `json:"body"`
	Files  []sdk.MaterialRef `json:"files"`
}

func manifest() (sdk.Manifest, error) {
	permissions := []string{sdk.PermissionBrowserNavigate, sdk.PermissionBrowserFileUpload, sdk.PermissionBrowserNetworkObserve, sdk.PermissionBrowserPageManage, sdk.PermissionStorageRead}
	input := sdk.Object(map[string]sdk.Schema{
		"origin": sdk.String("示例页面来源"), "title": sdk.String("标题"), "body": sdk.String("正文"),
		"files": {Type: "array", Items: &sdk.Schema{Type: "object", Properties: map[string]sdk.Schema{"artifact_ref": sdk.String("素材引用")}, Required: []string{"artifact_ref"}}},
	}, "origin", "title", "body", "files")
	output := sdk.Object(map[string]sdk.Schema{"state": sdk.String("结果状态"), "response_ref": sdk.String("响应引用"), "content_id": {Type: "object"}}, "state")
	node, err := sdk.Describe(sdk.NodeMetadata{Type: nodeType, Version: 1, DisplayName: "上传并提交（示例插件）", Description: "上传素材、填写示例表单并核验业务返回码；不是站点适配器", Category: "插件示例", SecurityLevel: "high_risk", Idempotency: "non_idempotent"}, input, output, sdk.NodePolicy{
		RequiredPermissions: permissions, Capabilities: permissions, Resources: sdk.NodeResources{ProfileLease: true, Runtime: true, ActivePage: true, ArtifactIO: true}, MaxAttempts: 1,
		Timeouts: sdk.NodeTimeout{Default: 2 * time.Minute, Minimum: time.Second, Maximum: 15 * time.Minute},
	})
	if err != nil {
		return sdk.Manifest{}, err
	}
	node.Schema.Ports = append(node.Schema.Ports, sdk.NodePort{Name: "unconfirmed", Kind: "branch", Description: "提交结果不确定，核实后再决定是否重试"}, sdk.NodePort{Name: "rejected", Kind: "branch", Description: "收到响应但业务返回码未通过"})
	return sdk.Manifest{ID: "examples.managed_publish", Name: "受控发布示例", Version: "0.1.0", Kind: sdk.KindAutomation, Entrypoint: "bin/plugin", Permissions: permissions, Nodes: []sdk.NodeDescriptor{node}}, nil
}

func handle(ctx context.Context, call sdk.NodeCall) (sdk.NodeResult, error) {
	var input publishInput
	if err := call.DecodeInput(&input); err != nil {
		return sdk.NodeResult{}, err
	}
	u, err := url.Parse(input.Origin)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || strings.TrimSpace(input.Origin) != input.Origin {
		return sdk.NodeResult{}, errors.New("origin must be an exact HTTP(S) origin without a path")
	}
	if err := call.Client.Navigate(ctx, input.Origin+"/compose"); err != nil {
		return sdk.NodeResult{}, err
	}
	watch := sdk.ResponseWatch{URLPattern: input.Origin + "/upload", Method: "POST", TimeoutMS: 60000}
	upload, err := call.Client.UploadFiles(ctx, sdk.UploadRequest{Selector: "#files", Origin: input.Origin, Files: input.Files, Multiple: len(input.Files) > 1, Watch: &watch, Checks: []sdk.ResponseCheck{{Name: "accepted", Path: "/responses/0/body/code", Equals: 0}}})
	if err != nil {
		return sdk.NodeResult{}, err
	}
	if upload.State != "response_received" || upload.Response == nil {
		return outcome("unconfirmed", nil), nil
	}
	if len(upload.Response.MissingChecks) > 0 {
		return outcome("unconfirmed", upload.Response), nil
	}
	if !acceptedResponse(upload.Response) {
		return outcome("rejected", upload.Response), nil
	}
	if err := call.Client.BindPage(ctx, sdk.PageBindingRequest{PageKey: "compose", Origin: input.Origin, GuardSelector: "#workspace", ExpectedText: "Example workspace"}); err != nil {
		return sdk.NodeResult{}, err
	}
	if err := call.Client.FillBoundPage(ctx, "compose", "#title", input.Title); err != nil {
		return sdk.NodeResult{}, err
	}
	if err := call.Client.FillBoundPage(ctx, "compose", "#body", input.Body); err != nil {
		return sdk.NodeResult{}, err
	}
	response, err := call.Client.WaitForBoundPageResponse(ctx, "compose", sdk.WaitResponseRequest{
		Watch: sdk.ResponseWatch{URLPattern: input.Origin + "/publish", Method: "POST", TimeoutMS: 30000}, Action: sdk.ResponseAction{Action: "click", Selector: "#publish"},
		Checks: []sdk.ResponseCheck{{Name: "accepted", Path: "/responses/0/body/code", Equals: 0}}, Fields: []sdk.ResponseField{{Name: "content_id", Path: "/responses/0/body/id"}},
	})
	if err != nil {
		return sdk.NodeResult{}, err
	}
	// Do not turn cleanup failure into an automatic resubmission. Run cleanup
	// also retires bindings; the publishing result remains authoritative here.
	_ = call.Client.ReleasePage(ctx, "compose")
	if response.Outcome != "response" {
		return outcome("unconfirmed", &response), nil
	}
	if len(response.MissingChecks) > 0 {
		return outcome("unconfirmed", &response), nil
	}
	if !acceptedResponse(&response) {
		return outcome("rejected", &response), nil
	}
	return outcome("confirmed", &response), nil
}

func acceptedResponse(response *sdk.ResponseResult) bool {
	return response.Status != nil && *response.Status >= 200 && *response.Status < 300 && response.Checks["accepted"]
}

func outcome(state string, response *sdk.ResponseResult) sdk.NodeResult {
	port := state
	if state == "confirmed" {
		port = "success"
	}
	output := map[string]any{"state": state}
	if response != nil {
		output["response_ref"] = response.SecretRef
		if value := response.Fields["content_id"]; value != nil {
			output["content_id"] = value
		}
	}
	return sdk.NodeResult{NextPort: port, Output: output}
}

func run() error {
	m, err := manifest()
	if err != nil {
		return err
	}
	if err = m.Validate(); err != nil {
		return err
	}
	if len(os.Args) == 2 && os.Args[1] == "--manifest" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(m)
	}
	server := sdk.NewServer()
	if err = server.Register(nodeType, 1, handle); err != nil {
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
