// Package scenarios is the source of workflows/*.json. It uses only the public
// SDK; placeholder sites/selectors/credentials must be replaced before running.
package scenarios

import (
	"fmt"
	wf "github.com/browser-rpa/anon-browser-sdk/workflow"
)

const origin = "https://example.test"
const sessionID = "session-REPLACE_ME"

func All() (map[string]wf.Workflow, error) {
	builders := legacyBuilders()
	builders["login-and-keep.json"] = wf.New("local", "人工登录并保持浏览器").Description("先替换环境 ID；检测到登录状态后保存会话，任务结束不关闭浏览器。").Add(
		wf.Start("start"), wf.N("environment", "environment.select", map[string]any{"environment_id": "environment-REPLACE_ME"}), wf.Open("open", false), wf.Navigate("login", origin+"/login"),
		wf.HumanWait("wait", wf.HumanWaitConfig{Mode: "auto", Title: "请在浏览器中完成登录", TimeoutMS: 600000, StableMS: 1000, Conditions: []wf.WaitCondition{{Kind: "element_visible", Selector: "[data-test=logged-in]"}}}),
		wf.KeepSession("keep", "演示常驻浏览器", "session.saved"), wf.End("end"), wf.End("not-confirmed"),
	).Then("start", "environment", "open", "login", "wait", "keep", "end").Link("wait", "timeout", "not-confirmed").Link("wait", "rejected", "not-confirmed")

	builders["reuse-session.json"] = wf.New("local", "复用已登录会话读取内容").Description("绑定会话需要在打开浏览器之前；忙碌时不会抢占。固定页面，不让外部参数控制 URL 或会话。").Add(
		wf.Start("start"), wf.BindSession("bind", sessionID), wf.Open("open", false), wf.Navigate("page", origin+"/workspace"), wf.ReadText("read", "[data-test=content]", "page.content"),
		wf.Webhook("notify", "https://callback.example.invalid/content", "secret://callbacks/signing", map[string]any{"content": wf.Ref("page.content")}), wf.End("end"),
	).Then("start", "bind", "open", "page", "read", "notify", "end")

	builders["session-script-listener.json"] = wf.New("local", "为常驻会话注册消息监听").Description("教学自定义事件入口，不是任何站点的私信协议；先配置连接器及 start_run 触发器，再注册脚本。").Add(
		wf.Start("start"), wf.BindSession("bind", sessionID), wf.Open("open", false),
		wf.RegisterScript("listen", wf.ScriptConfig{Key: "message-listener", Origin: origin, RunAt: "document_start", Scope: "session", ConnectorID: "connector-REPLACE_ME", ConnectorRevision: 1,
			Source: `window.addEventListener('example:message', e => {
 const m = e.detail;
 if (!m || typeof m.id !== 'string' || typeof m.text !== 'string' || m.id.length > 128 || m.text.length > 8000) return;
 bridge.emit(m.id, {text: m.text});
});`}), wf.Navigate("page", origin+"/messages"), wf.End("end"),
	).Then("start", "bind", "open", "listen", "page", "end")

	builders["external-command.json"] = wf.New("local", "等待外部指令并确认处理").Description("signal_run 入口绑定该运行；这里只演示领取和确认，不假装已完成站点发送。").Add(
		wf.Start("start"), wf.WaitSignal("wait", wf.SignalWaitConfig{SignalName: "chat.reply", TimeoutMS: 600000, OutputVariable: "command.receipt", Fields: []wf.SignalField{{Name: "text", Type: wf.WorkflowInputString, Required: true, Variable: "command.text"}}}),
		wf.HumanWait("confirm", wf.HumanWaitConfig{Mode: "manual", Title: "完成实际业务操作并核验结果后确认", TimeoutMS: 300000}),
		wf.CompleteSignal("done", wf.Ref("command.receipt"), "completed"), wf.CompleteSignal("failed", wf.Ref("command.receipt"), "failed"), wf.End("end"), wf.End("no-command"),
	).Then("start", "wait", "confirm", "done", "end").Link("wait", "timeout", "no-command").Link("confirm", "timeout", "failed").Link("confirm", "rejected", "failed").Link("failed", "success", "end")

	builders["reply-request.json"] = wf.New("local", "上报事件并等待外部回复").Description("示例固定事件需换成可信采集结果；reply_request 触发器需先配置并启用；人工确认仅演示业务核验位置。").Add(
		wf.Start("start"), wf.RequestReply("request", wf.ReplyRequestConfig{SignalName: "chat.reply", MessageKey: "example-message-1", Context: map[string]any{"conversation_id": "example-conversation"}, OutputVariable: "reply.request"}),
		wf.Webhook("notify", "https://callback.example.invalid/messages", "secret://callbacks/signing", wf.Ref("reply.request")),
		wf.WaitSignal("wait", wf.SignalWaitConfig{SignalName: "chat.reply", OutputVariable: "reply.receipt", Fields: []wf.SignalField{{Name: "text", Type: wf.WorkflowInputString, Required: true, Variable: "reply.text"}}}),
		wf.HumanWait("confirm", wf.HumanWaitConfig{Mode: "manual", Title: "核验业务操作结果后确认"}), wf.CompleteSignal("done", wf.Ref("reply.receipt"), "completed"), wf.CompleteSignal("failed", wf.Ref("reply.receipt"), "failed"), wf.End("end"), wf.End("expired"),
	).Then("start", "request", "notify", "wait", "confirm", "done", "end").Link("wait", "timeout", "expired").Link("confirm", "timeout", "failed").Link("confirm", "rejected", "failed").Link("failed", "success", "end")

	watch := wf.ResponseWatch{URLPattern: origin + "/api/upload", Method: "POST", Selection: "last", QuietMS: 1000, TimeoutMS: 30000, BodyMode: "json"}
	builders["upload-and-confirm.json"] = wf.New("local", "上传素材并核验页面结果").Description("替换素材引用、会话及站点字段；收到响应不等于业务成功。最后一个响应以 1 秒静默为收集边界。").Add(
		wf.Start("start"), wf.BindSession("bind", sessionID), wf.Open("open", false), wf.Navigate("page", origin+"/publish"),
		wf.Upload("upload", wf.UploadConfig{Selector: "input[type=file]", Origin: origin, Files: []wf.MaterialRef{wf.Material("artifact://material-00000000000000000000000000000000")}, Watch: &watch, OutputVariable: "upload.result", ResponseOutputVariable: "upload.response", Extract: []wf.ExtractField{{Path: "/responses/0/body/id", Variable: "upload.asset_id"}}}),
		wf.HumanWait("verify", wf.HumanWaitConfig{Mode: "auto", TimeoutMS: 30000, Conditions: []wf.WaitCondition{{Kind: "element_visible", Selector: "[data-test=upload-complete]"}}}), wf.End("end"), wf.End("not-confirmed"),
	).Then("start", "bind", "open", "page", "upload", "verify", "end").Link("upload", "timeout", "not-confirmed").Link("upload", "failed", "not-confirmed").Link("upload", "condition_unmet", "not-confirmed").Link("verify", "timeout", "not-confirmed").Link("verify", "rejected", "not-confirmed")

	builders["wait-response.json"] = wf.New("local", "执行动作并等待接口响应").Add(
		wf.Start("start"), wf.BindSession("bind", sessionID), wf.Open("open", false), wf.Navigate("page", origin+"/workspace"),
		wf.WaitResponse("response", wf.WaitResponseConfig{Watch: wf.ResponseWatch{URLPattern: origin + "/api/content", Method: "GET", Selection: "first", BodyMode: "json"}, Action: wf.ResponseAction{Action: "click", Selector: "[data-test=load]"}, OutputVariable: "network.response"}), wf.End("end"), wf.End("not-confirmed"),
	).Then("start", "bind", "open", "page", "response", "end").Link("response", "timeout", "not-confirmed").Link("response", "failed", "not-confirmed").Link("response", "condition_unmet", "not-confirmed")

	builders["managed-download.json"] = wf.New("local", "下载文件到素材库").Add(
		wf.Start("start"), wf.BindSession("bind", sessionID), wf.Open("open", false), wf.Navigate("page", origin+"/reports"),
		wf.Download("download", wf.DownloadConfig{Selector: "[data-test=download]", Origin: origin, URLPattern: origin + "/report.csv", FilenamePattern: "report*.csv", MaxBytes: 32 << 20, OutputVariable: "download.files"}), wf.End("end"),
	).Then("start", "bind", "open", "page", "download", "end")

	sub, err := wf.New("local", "共享子流程").Add(wf.Start("start"), wf.SetVariable("set", "result", "ready"), wf.End("end")).Then("start", "set", "end").Returns(map[string]any{"status": wf.Ref("result")}).Build()
	if err != nil {
		return nil, err
	}
	builders["shared-subflow.json"] = wf.New("local", "同一子流程调用两次").Subflow(wf.SubflowDefinition{ID: "prepare", Name: "准备步骤", Graph: sub.Graph}).Add(
		wf.Start("start"), wf.Call("first", wf.CallConfig{SubflowID: "prepare", Outputs: map[string]string{"status": "first.status"}}), wf.Call("again", wf.CallConfig{SubflowID: "prepare", Outputs: map[string]string{"status": "again.status"}}), wf.End("end"),
	).Then("start", "first", "again", "end")
	result := map[string]wf.Workflow{}
	for name, builder := range builders {
		value, err := builder.Build()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		result[name] = value
	}
	return result, nil
}
