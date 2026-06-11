# Week 7 任务说明：aria2 下载任务、进度、停止与重试

## 本周定位

本周完成服务端下载任务闭环。OpenBridge 负责把 OpenList 文件解析为直链，再提交给运行在服务端电脑上的 aria2，并持续同步任务状态。

## 本周目标

- 支持创建 aria2 下载任务。
- 保存任务与 aria2 GID 映射。
- 支持任务列表、任务详情和状态筛选。
- 支持每秒刷新活跃任务状态。
- 支持进度条、下载速度和预测进度。
- 支持停止单个任务和批量停止任务。
- 支持失败或手动停止任务重试。
- 支持打开已下载文件、打开所在文件夹和删除本地文件。

## 任务拆分

### 后端

- 封装 aria2 JSON-RPC Client。
- 实现 `POST /api/v1/download/tasks`。
- 实现 `GET /api/v1/download/tasks/:id`。
- 实现 `GET /api/v1/download/aria2-status`。
- 实现 `POST /api/v1/download/tasks/:id/stop`。
- 实现 `POST /api/v1/download/tasks/stop`。
- 实现 `POST /api/v1/download/tasks/:id/retry`。
- 实现 `POST /api/v1/download/tasks/:id/open`。
- 实现 `POST /api/v1/download/tasks/:id/open-location`。
- 实现 `POST /api/v1/download/tasks/:id/delete-file`。
- 任务实体保存 `Progress`、`CompletedLength`、`TotalLength`、`DownloadSpeed`、`RetryCount` 等字段。
- 修复特殊文件名打开所在文件夹失败问题。
- 删除本地文件时保留任务记录，并将状态更新为 `deleted`。

### 前端

- 下载任务页支持创建任务。
- 源路径支持 OpenList 文件选择器。
- 目标目录支持本机目录选择。
- 任务列表支持筛选：全部、等待中、下载中、暂停、停止、文件已删除、失败、完成。
- 任务列表支持排序和多选。
- 任务行展示进度条、下载速度、已下载大小和总大小。
- 任务行支持停止、删除文件、清除记录、重试、打开文件、打开所在文件夹。
- 任务详情展示完整字段和操作按钮。
- 存在活跃任务时默认每秒刷新。

### 状态规则

- 等待中、下载中、暂停中任务可以停止。
- 失败和已停止任务可以重试。
- 已完成、已停止和失败任务可以删除本地文件。
- 清除记录不删除本地文件。

## 验收标准

- 从 OpenList 页面和任务页面都能创建 aria2 下载任务。
- 任务进度、速度和状态能持续更新。
- 停止任务后可重试。
- 失败任务可重试。
- 已完成任务可打开文件或所在文件夹。
- 删除文件后任务记录仍存在且状态为文件已删除。
- 批量停止能返回成功列表和失败列表。

## 交付物

- aria2 Client。
- DownloadTask API。
- 下载任务页面。
- 任务状态同步逻辑。
