# Week 8 任务说明：WebDAV Mount 代理与 rclone 挂载配置

## 本周定位

本周完成本地盘符挂载链路。OpenBridge 通过 WebDAV 代理复用 OpenList 文件能力，同时用 OpenBridge Mount 层配额改写客户端看到的容量，再通过 rclone 管理本地挂载。

## 本周目标

- 支持按 Mount 暴露 WebDAV 根。
- WebDAV 文件能力复用 OpenList。
- WebDAV 容量展示使用 OpenBridge Mount 配额。
- 支持 rclone 配置保存、写入、挂载、停止挂载和删除。
- 支持普通、Union、Combine 三种挂载方式。
- 支持 rclone 路径从设置页保存到 `.env`。
- 修复 OpenList 端口更换后 rclone 仍引用旧配置的问题。

## 任务拆分

### WebDAV 代理

- 实现 `/api/v1/webdav/mounts/:id`。
- 根据 Mount ID 找到对应 `mount_path`。
- 将文件操作代理到 `{OPENLIST_BASE_URL}/dav{mount.mount_path}`。
- 透传客户端 Authorization 给 OpenList。
- 改写 `PROPFIND` 中的 `href`、`Location`、`Content-Location`。
- 改写 `quota-used-bytes` 和 `quota-available-bytes`。
- 支持常见 WebDAV 方法：`OPTIONS`、`PROPFIND`、`GET`、`HEAD`、`PUT`、`DELETE`、`MKCOL`、`MOVE`、`COPY`。

### rclone 后端

- 定义 RcloneProfile 实体。
- RcloneProfile 按 OpenList Base URL 隔离。
- 保存配置名、挂载方式、Mount IDs、用户名、暗文密码和目标路径。
- 实现配置列表、新增、编辑、删除。
- 实现写入 rclone 配置。
- 实现启动 `rclone mount`。
- 实现停止对应挂载进程。
- Mount 被删除后，旧 rclone 配置需要刷新状态或阻止继续挂载。

### 前端

- 新增 Rclone 页面。
- 新增配置弹窗。
- 支持选择 Mount。
- 支持普通、Union、Combine 三种模式。
- 支持目标盘符或本地目录输入。
- 支持本机路径选择器。
- 配置卡片展示写入、挂载、停止挂载、删除和复制命令。

## 验收标准

- 单个 Mount 可以通过 WebDAV 地址访问。
- rclone 可把 Mount 挂载为本地盘符。
- 容量展示尽量贴近 OpenBridge Mount 配额。
- Union/Combine 配置能保留多个 Mount 的识别信息。
- 停止挂载后本地盘符释放。
- 删除配置后 rclone 中旧配置不继续残留为可用状态。
- 更换 OpenList 端口后旧源 rclone 配置不会串到新源。

## 交付物

- WebDAV Mount 代理。
- Rclone Profile API。
- Rclone 配置页面。
- WebDAV Mount 使用说明。
