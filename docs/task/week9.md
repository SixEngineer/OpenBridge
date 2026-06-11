# Week 9 任务说明：设置中心、主机资源、性能与界面体验

## 本周定位

本周完成控制台的运行时管理能力。目标是让用户不再手动改命令和配置文件，而是在前端完成 OpenList、aria2、rclone、缓存、登录策略和服务控制配置。

## 本周目标

- 设置页拆分为清晰的配置卡片。
- 支持 aria2 RPC、aria2 路径和自动启动。
- 支持 rclone 路径保存。
- 支持 OpenList Base URL 保存并触发重新登录。
- 支持本地路径和文件选择器。
- 支持重启和退出 OpenBridge。
- 支持启动自动打开浏览器。
- 支持主机资源和带宽监控。
- 优化前端动画、主题和响应式布局。

## 任务拆分

### 后端设置

- 实现 `GET /api/v1/settings`。
- 实现 `PUT /api/v1/settings/openlist`。
- 实现 `PUT /api/v1/settings/aria2`。
- 实现 `PUT /api/v1/settings/rclone`。
- 实现 `PUT /api/v1/settings/session`。
- 实现 `PUT /api/v1/settings/app`。
- 实现 `PUT /api/v1/settings/filetree`。
- 设置写入 `.env`，并更新运行时配置。
- 设置返回 `app_version`，前端底部显示 `OpenBridge vX.X`。

### 系统能力

- 实现 `POST /api/v1/system/pick-path`。
- 实现 `GET /api/v1/system/metrics`。
- 实现 `POST /api/v1/system/restart`。
- 实现 `POST /api/v1/system/exit`。
- 主机资源展示整机 CPU、内存、磁盘和 OpenBridge 进程占用。
- 带宽展示上行、下行、OpenBridge、OpenList、aria2 和 rclone 相关指标。
- 服务启动时根据配置自动打开浏览器。
- 服务启动时根据配置自动拉起 aria2。

### 前端体验

- 设置页分为用户信息、aria2、默认下载目录、其他设置、服务控制、FILETREE、Rclone、OpenList、重置用户数据。
- rclone 设置单独成框。
- aria2 路径、自动启动、rclone 路径能正确保存。
- 默认下载目录和路径输入均支持文件或目录选择器。
- 仪表盘展示总空间、Provider 用量、主机资源、系统健康和快捷操作。
- 顶部栏展示实时上下行速度。
- 支持用户选择是否开启动画效果。
- 修复仪表盘、服务商、配额页面大面积空白和黑夜模式可读性问题。

## 验收标准

- 设置页保存后重启仍生效。
- aria2 能自动启动。
- rclone 路径能写入 `.env` 并用于 rclone 操作。
- 点击重启服务和退出服务行为明确。
- 仪表盘资源占用每秒刷新且无明显卡顿。
- 前端在桌面和移动端都能正常渲染。

## 交付物

- Settings API。
- System API。
- 设置页面。
- 仪表盘和顶部栏指标。
- 本机路径选择器。
