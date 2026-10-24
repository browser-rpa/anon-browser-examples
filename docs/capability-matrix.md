# 公开能力示例矩阵

本页以公开客户端 `GET /api/rpa/node-catalog` 的 Node Schema 为准，帮助开发者从“我要做什么”找到对应示例。节点配置必须以客户端返回的 Catalog revision 和 Schema 为准；示例不复制私有实现。

## 浏览器节点

| 节点 | 示例 |
| --- | --- |
| `browser.open`、`browser.close`、`browser.navigate`、`browser.reload` | `workflows/browser-basics.json` |
| `browser.wait`、`browser.click`、`browser.double_click`、`browser.right_click`、`browser.hover` | `workflows/browser-basics.json` |
| `browser.type`、`browser.clear`、`browser.press`、`browser.select`、`browser.check` | `workflows/browser-basics.json` |
| `browser.read_text`、`browser.read_attribute`、`browser.read_title`、`browser.read_url`、`browser.inspect` | `workflows/browser-basics.json` |
| `browser.scroll`、`browser.screenshot` | `workflows/browser-basics.json`、`workflows/artifacts.json` |
| `browser.new_tab`、`browser.switch_tab`、`browser.close_tab`、`browser.back`、`browser.forward` | `workflows/tabs.json` |
| `browser.evaluate` | `workflows/page-script-and-network.json` |
| `browser.network_rule`、`browser.network_clear` | `workflows/page-script-and-network.json` |

## 数据、会话与集成节点

| 节点 | 示例 |
| --- | --- |
| `data.set`、`data.get` | `workflows/control-flow.json`、`workflows/artifacts.json` |
| `browser.cookies_set`、`browser.cookies_get` | `workflows/session-storage.json` |
| `browser.storage_set`、`browser.storage_get` | `workflows/session-storage.json` |
| `http.request` | `workflows/page-script-and-network.json` |
| `artifact.create` | `workflows/artifacts.json` |

## 流程控制节点

`flow.delay`、`flow.if` 和 `flow.switch` 见 `workflows/control-flow.json`；`flow.loop`、`flow.for_each`、`flow.break` 和 `flow.continue` 见 `workflows/loops.json`；`flow.try` 和 `flow.retry` 见 `workflows/error-handling.json`。循环和分支的端口连接是 Graph DSL 的一部分，建议先在编辑器中打开示例，再按业务修改。

## 环境和后台任务

`environment.provision`、`environment.select`、`environment.retain` 和 `environment.discard` 见 `workflows/environments.json`、`workflows/select-environment.json` 和 `workflows/discard-environment.json`。后台任务使用 `background.start`、`background.wait`、`background.stop`，元素观察见 `workflows/background-watch.json`，URL 观察见 `workflows/background-url-watch.json`；当前公开示例不包含 WebSocket 监听。

## HTTP API

- `scripts/health.sh`：检查本地客户端是否运行；
- `scripts/profile-lifecycle.sh`：创建、查询和删除 Profile；
- `scripts/workflow-lifecycle.sh`：创建、读取、校验、发布、运行、读取事件和下载产物；
- `scripts/import-plugin.sh`：导入自动化插件；
- `scripts/import-extension.sh`：导入本地浏览器扩展；
- `scripts/curl-workflow.sh`：只创建一个工作流，适合快速试验。

本地 API 的完整字段以公开文档仓库中的 OpenAPI 文件和客户端实际 Catalog 为准。
