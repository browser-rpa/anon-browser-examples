# anon-browser examples

这里是 anon-browser 公开开发者示例仓库，只包含可以在公开客户端 API、RPA 节点 Schema 和自动化插件契约上运行的示例。示例面向已经下载并启动 anon-browser 桌面客户端的本地开发者，不包含服务端部署或私有实现。

## 目录

- `plugins/echo-node`：最小自动化插件，导出一个把输入写入输出的 RPA 节点。
- `plugins/capability-node`：Go 编写的 capability 调用示例，演示读取页面、读写工作流变量和发起 HTTP 请求。
- `workflows/`：按能力拆分的工作流 Graph，包括浏览器操作、标签页、变量和流程控制、会话数据、网络规则、后台观察、文件产物和临时环境。
- `docs/capability-matrix.md`：公开节点、API 和示例文件的完整对应表。
- `scripts/package-plugin.sh`：在本机编译并打包自动化插件。
- `scripts/`：通过本地客户端 HTTP API 调用健康检查、Profile、Runtime、工作流、Run、计划、插件、扩展和 Artifact 的脚本。

## 要求

- Go 1.22+，用于构建示例插件；
- 已下载并启动 anon-browser 桌面客户端；
- 本地 API 地址默认为 `http://127.0.0.1:8787`，可通过 `ANON_BROWSER_API` 覆盖。

## 打包并导入插件

```bash
./scripts/package-plugin.sh echo-node
# capability 示例
./scripts/package-plugin.sh capability-node
```

脚本会生成 `dist/echo-node-<os>-<arch>.abp`。随后可以在桌面端“扩展中心 → 自动化插件”导入，或执行：

```bash
curl -X POST "$ANON_BROWSER_API/api/plugins/automation/import" \
  -F 'file=@dist/echo-node-darwin-arm64.abp'
```

导入后重新打开 RPA 编辑器，节点 `examples.echo` 或 `examples.capability.*` 会出现在左侧节点库。

## Go 插件如何调用 Core 能力

自动化插件是一个由 Core 启动的 Go 可执行文件。插件不直接连接 CDP、Profile
目录或数据库，而是通过 stdin/stdout 传递 JSON。Core 先发送一次节点调用，插件
需要能力时向 stdout 发送一条 `capability_call`，再从 stdin 读取对应的
`capability_result`，最后发送 `ok: true` 的节点结果。

例如读取当前页面标题：

```json
{"type":"capability_call","id":"call-1","capability":"browser","action":"read_title","arguments":{}}
```

Core 返回：

```json
{"type":"capability_result","id":"call-1","ok":true,"data":"页面标题"}
```

插件随后返回：

```json
{"ok":true,"result":{"next_port":"success","output":{"value":"页面标题"},"variables":{"page.title":"页面标题"}}}
```

`capability-node` 包含完整的 Go 实现。它演示了以下调用：

| capability | action | 用途 |
| --- | --- | --- |
| `browser` | `read_title`、`read_url` | 读取当前页面 |
| `variables` | `get`、`set` | 读取或写入当前 Run 变量 |
| `http` | `request` | 通过 Core 发起外部 HTTP 请求 |

插件必须在 Manifest 的 `permissions` 中声明对应权限，节点 Policy 的
`required_permissions` 也必须声明同一权限；否则调用会被 Core 拒绝。详细的能力边界见
[公开 capability 文档](https://browser-rpa.github.io/anon-browser-docs/plugins/capabilities)。

## 工作流示例

所有工作流文件都使用公开的 Graph DSL，可以直接在客户端的 RPA 编辑器导入，或通过 API 创建：

```bash
./scripts/curl-workflow.sh workflows/browser-basics.json
```

示例中的 `https://example.test`、`profile-...` 和 `environment-...` 是占位值。运行前请替换为自己的测试网址、Profile ID 或环境 ID；涉及 Cookie、Storage、JavaScript、请求改写和外部 HTTP 的工作流需要在客户端中显式批准相应权限。

### 能力覆盖

| 能力 | 示例 |
| --- | --- |
| 浏览器启动、导航、等待、点击、输入、按键、滚动、读取文本/属性/标题/URL、截图 | `workflows/browser-basics.json` |
| 新建、切换、关闭标签页，以及前进、后退、刷新 | `workflows/tabs.json` |
| 变量、条件、分支、延迟 | `workflows/control-flow.json` |
| 循环、遍历、异常处理和重试 | `workflows/loops.json`、`workflows/error-handling.json` |
| Cookie、localStorage/sessionStorage 写入和读取 | `workflows/session-storage.json` |
| JavaScript、请求放行/拦截/Header 修改/Body 替换 | `workflows/page-script-and-network.json` |
| 后台元素观察、URL 观察和等待/停止后台任务 | `workflows/background-watch.json`、`workflows/background-url-watch.json`、`workflows/background-stop.json` |
| 文本、JSON、CSV、HTML 文件产物和截图产物 | `workflows/artifacts.json` |
| 临时环境创建、已有环境选择、保存和丢弃 | `workflows/environments.json`、`workflows/select-environment.json`、`workflows/discard-environment.json` |
| HTTP API 创建、发布、运行、调试、事件和 Artifact 下载 | `scripts/workflow-lifecycle.sh`、`scripts/download-artifact.sh` |
| Runtime 查询、启动和停止 | `scripts/runtime-lifecycle.sh` |
| Cron 计划创建 | `scripts/schedule-workflow.sh` |
| Profile、浏览器扩展和自动化插件管理 | `scripts/profile-lifecycle.sh`、`scripts/import-extension.sh`、`scripts/import-plugin.sh` |

这些文件展示配置方式和节点连接方式；页面选择器、Cookie 值、外部 URL 和权限应按你的测试目标替换，不要把真实 Secret 提交到仓库。

## 公开边界

示例不会依赖私有仓库、内部数据库、Profile 数据目录、原始 CDP 地址、Gateway 配置或服务端部署逻辑。需要浏览器能力时，应通过插件 Host 的 capability 协议请求，不要自行连接浏览器调试端口。
