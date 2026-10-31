# 上传并提交示例插件

用于验证一个业务插件能否独立完成：打开表单 → 上传授权素材 → 核验上传返回码 → 绑定页面 → 填写 → 提交 → 核验发布返回码。不是小红书适配器，不含登录凭据，不调用任何真实平台。

需要支持受控浏览器接口的 SDK 和客户端。示例已固定到公开 SDK 预览版本，可用 `GOWORK=off` 独立构建；不要提交机器相关 replace。

## 工作流配置

1. 工作流打开浏览器，或绑定空闲会话后打开。
2. 插入 `examples.managed_publish@1`，填写来源、标题、正文、素材引用。
3. 分别连接 `success`、`unconfirmed`、`rejected`。失败/未知结果不得自动重试发布。

节点不创建或关闭会话。素材引用必须在本次工作流配置/运行输入中，并在运行创建时通过授权。

如果文件来自 CDN，先在素材库点击「链接导入」，填写公开可下载的 HTTPS 链接与文件名，导入后选择其素材引用。不把 URL、本机路径、浏览器 Cookie 或管理凭证作为插件的文件输入。链接导入只负责准备素材，不会自动运行本插件或发布内容。

## 示例站点契约

- `GET /compose`：页面有 `#workspace`（文本 `Example workspace`）、`#files` 文件输入框、`#title`、`#body`、`#publish`。
- 选择文件后页面发起 `POST /upload`，完成响应 JSON 为 `{"code":0}`；页面应读取响应正文。
- 点击发布后页面发起 `POST /publish`，完成响应 JSON 为 `{"code":0,"id":"..."}`；页面应读取响应正文。

节点使用宿主内的 `Checks` 判断 code，不把响应正文/Cookie 暴露给日志。内容 ID 以加密引用返回。HTTP 200 或单纯点击成功不直接当作发布成功。

真实平台可能需要额外等待素材处理、不同的富文本交互或 trusted 输入；不能将这个教学契约视为已完成某个平台适配。

## 构建

```bash
go run ./plugins/managed-publish --manifest
bash scripts/package-plugin.sh managed-publish
```

Manifest 来源为 Go，不手写。`scripts/check-generated.sh` 同时检查其生成物。
