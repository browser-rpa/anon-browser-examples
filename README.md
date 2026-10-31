# anon-browser examples

anon-browser 插件与工作流开发示例，基于 [anon-browser-sdk](https://github.com/browser-rpa/anon-browser-sdk)，演示节点定义、能力调用、工作流编排和客户端 API 集成。

## 本地准备

Go 1.24+。`go.mod` 固定了公开 SDK 的预览提交版本，但本分支的受控发布插件和会话准备示例使用了尚未发布的新接口，暂时不能仅靠该版本完成全量编译。SDK 尚未发布稳定版本标签。

当前联调需将本分支配套的 `anon-browser-sdk` 放在相邻目录。在仓库根目录创建本地工作区（不要提交 `go.work` / `go.work.sum`）：

```bash
# 已有 go.work 时用 go work use . ../anon-browser-sdk 补充模块。
go work init . ../anon-browser-sdk
```

SDK 新版本发布并更新 `go.mod` 后，再验证不依赖工作区的独立构建。

示例编译、生成和测试都不需要运行桌面应用；真正执行需要已启动的 anon-browser 客户端。API 默认为 `http://127.0.0.1:8787`，可通过 `ANON_BROWSER_API` 覆盖。

## 源码与生成物

- `plugins/echo-node/main.go`：共享 SDK 协议、Go 节点描述和最小 handler。
- `plugins/capability-node/main.go`：类型化读取页面、变量操作、HTTP 调用；不再手写协议结构体、ID 或 stdin/stdout 往返。
- `plugins/managed-publish/`：上传素材、填写表单、提交并核验业务返回码的闭环示例；新增 SDK 接口尚未发布，当前使用 `go.work` 联调，详见该目录说明。
- `scenarios/scenarios.go`：新增能力的 Go 编排示例。
- `scenarios/legacy.go`：原基础示例的 Go 源码，也已按当前 Catalog 修正。
- `workflows/*.json`、`plugins/*/manifest.json`：生成后的公开交换文件，不直接编辑。
- `scripts/`：导入、查询和运行的 HTTP 示例。

```bash
go test ./...
go run ./cmd/generate-workflows
go run ./cmd/generate-workflows --check
bash scripts/check-generated.sh
```

Manifest 的来源也是 Go。更新插件定义后生成：

```bash
go run ./plugins/echo-node --manifest > plugins/echo-node/manifest.json
go run ./plugins/capability-node --manifest > plugins/capability-node/manifest.json
go run ./plugins/managed-publish --manifest > plugins/managed-publish/manifest.json
```

## 插件如何开发

```go
server := plugin.NewServer()
err := server.Register("acme.page_title", 1, func(ctx context.Context, call plugin.NodeCall) (plugin.NodeResult, error) {
    title, err := call.Client.ReadTitle(ctx)
    if err != nil { return plugin.NodeResult{}, err }
    return plugin.NodeResult{NextPort: "success", Output: map[string]any{"value": title}}, nil
})
if err != nil { return err }
return server.Serve(context.Background(), os.Stdin, os.Stdout)
```

完整的 imports、main、Manifest 和权限声明见各插件目录。通过 `plugin.Describe`、`plugin.Object`、`plugin.String` 用 Go 描述节点；使用 `time.Second`，不需要手算 JSON 中的纳秒值。

```bash
./scripts/package-plugin.sh echo-node
./scripts/package-plugin.sh capability-node
# 支持交叉编译；Manifest 生成仍在构建主机执行：
GOOS=windows GOARCH=amd64 ./scripts/package-plugin.sh echo-node
```

产物为 `dist/<name>-<os>-<arch>.abp`。从“扩展中心 → 自动化插件”导入，安装/升级后重启客户端使节点重新注册。打包脚本直接从 Go 生成 Manifest，不依赖可能过期的 JSON 快照。

插件能力须由 Manifest 和节点 Policy 声明，运行时还受宿主授权。声明受限权限的节点应为 `sensitive` 或 `high_risk`，不能标为 `safe`。独立进程不是不可信代码沙箱，只运行可信插件。

`capability-node` 的每个节点都支持 `output_variable`，所以除主操作权限外，还声明 `variables.write`。只返回 `Output` 的节点不需要这个权限。HTTP 节点使用 `high_risk`；运行时不能借用同包其他节点的权限。Manifest 生成过程会校验声明并指出出错节点与字段。

## 工作流用 Go 编排

```go
flow, err := workflow.New("local", "会话任务").Add(
    workflow.Start("start"),
    workflow.BindSession("bind", "session-REPLACE_ME"),
    workflow.Open("open", false),
    workflow.Navigate("page", "https://example.test/workspace"),
    workflow.End("end"),
).Then("start", "bind", "open", "page", "end").Build()
```

`Then` 连接 success；其他分支必须显式 `Link`。变量用 `workflow.Ref("name")`；复杂节点用类型化 Config；尚未封装的节点可用 `N(id, type, config)`，无需手写 JSON 字符串。生成后仍要通过目标客户端的 Catalog 校验和发布。

```bash
./scripts/curl-workflow.sh workflows/login-and-keep.json
```

### 新增能力示例

| 文件 | 演示与前置条件 |
| --- | --- |
| `ensure-session.json` | 按固定标识首次创建或复用浏览器，再检测登录；不可用时停止，不自动替换会话 |
| `login-and-keep.json` | 选已有环境 → 打开 → 手动登录、自动检测完成 → 保持会话；首次先创建环境，或用环境创建节点替换选择节点 |
| `reuse-session.json` | 绑定已保存会话 → 读取固定页面 → Webhook 回传业务内容；替换 session ID、选择器和回调 |
| `session-script-listener.json` | 绑定会话 → 注册 document_start 脚本 → 导航；先配置事件连接器及启用的 start_run 触发器 |
| `external-command.json` | 等待 signal_run 指令 → 业务结果核验 → 完成/失败回执 |
| `reply-request.json` | 创建消息处理请求 → Webhook 通知 → 等待回复 → 核验 → 完成回执；先建 reply_request 入口 |
| `upload-and-confirm.json` | 从素材库交付文件，监听上传响应并检测页面完成状态；last 选择规则有显式静默边界 |
| `wait-response.json` | 先监听再点击，保存第一条匹配响应 |
| `managed-download.json` | 先监听再点击下载，完成后保存到素材库并得到引用 |
| `shared-subflow.json` | 在根图定义一次子流程，两次调用，分别映射结果 |

所有网址、选择器、连接器、会话 ID、Secret 和素材引用均为占位值，结构校验通过不代表对某个真实站点已适配。监听示例只接收教学自定义事件，不宣称已实现小红书私信监听。外部指令示例用人工核验标出真实业务操作的位置，不会伪装已发送消息。

### 外部调用与文件边界

- `start_run` 创建新运行；可在固定流程内绑定一个空闲的已有会话。`signal_run` 投递给绑定的活动运行，由 `signal.wait` 在安全点领取，不会任意跳转画布。
- `reply_request` 只回复已上报且未过期的处理请求；`signal.complete` 应在业务证据核验后执行。受理、领取消费、HTTP 200 均不等于发布/发送成功。
- 页面脚本事件进入 Connector Inbox，经固定版本触发器启动独立 Run，不是自动恢复原 Run 的等待节点。外部通知再由该工作流的 Webhook 节点发出。
- Webhook 回调使用固定 HTTPS 地址和签名 Secret 引用。事件、响应、Cookie 等敏感数据不要直接放入回调或日志。
- 上传只接受项目内素材引用，不接收本机路径或 CDN URL。先上传素材库或通过受控网页下载入库，再传 `artifact_ref`。
- 响应监听支持 first/last/nth/all/first_match；last/all 必须配置数量、时间窗或静默边界。响应正文/Header/Cookie 保存在受控加密引用中，字段提取继续遵守宿主敏感数据规则。

原浏览器、标签页、变量、循环、后台观察、环境和产物示例仍保留，详见 [能力矩阵](docs/capability-matrix.md)。双击/右键使用 `browser.click` 的 `mode`，不要再使用旧独立节点类型。

## 验证边界

`go test ./...` 检查示例包能否编译；`scripts/check-generated.sh` 检查 Go 定义与生成的 Manifest、工作流 JSON 是否一致。实际运行前，请在目标版本的客户端中校验工作流并确认所需权限。示例中的网址、选择器和引用需按使用场景替换，不包含站点凭据或实际业务数据。
