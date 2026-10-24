# anon-browser examples

这里是 anon-browser 公开开发者示例仓库，只包含可以在公开客户端 API 和自动化插件契约上运行的示例。

## 目录

- `plugins/echo-node`：最小自动化插件，导出一个把输入写入输出的 RPA 节点。
- `workflows/google-search.json`：使用内置节点打开搜索页面的工作流 Graph。
- `scripts/package-plugin.sh`：在本机编译并打包自动化插件。
- `scripts/curl-workflow.sh`：通过本地客户端 HTTP API 创建和查询工作流的示例。

## 要求

- Go 1.22+，用于构建示例插件；
- 已下载并启动 anon-browser 桌面客户端；
- 本地 API 地址默认为 `http://127.0.0.1:8787`，可通过 `ANON_BROWSER_API` 覆盖。

## 打包并导入插件

```bash
./scripts/package-plugin.sh
```

脚本会生成 `dist/echo-node-<os>-<arch>.abp`。随后可以在桌面端“扩展中心 → 自动化插件”导入，或执行：

```bash
curl -X POST "$ANON_BROWSER_API/api/plugins/automation/import" \
  -F 'file=@dist/echo-node-darwin-arm64.abp'
```

导入后重新打开 RPA 编辑器，节点 `examples.echo` 会出现在左侧节点库。

## 公开边界

示例不会依赖私有仓库、内部数据库、Profile 数据目录、原始 CDP 地址、Gateway 配置或服务端部署逻辑。需要浏览器能力时，应通过插件 Host 的 capability 协议请求，不要自行连接浏览器调试端口。
