# Week 5 任务说明：OpenList 文件浏览、用户根目录与文件管理

## 本周定位

本周把 OpenList 文件能力接入 OpenBridge，让用户能在控制台中浏览自己可见的 OpenList 根目录，并补齐基础文件管理能力。

## 本周目标

- 支持按当前 OpenList 用户根目录浏览文件。
- 支持 OpenList 文件列表分页加载，避免只显示前 50 个文件。
- 支持文件详情读取。
- 支持删除、重命名、复制、剪切和粘贴。
- 支持文件类型图标和多选复选框。
- 支持下载任务和路径输入场景复用文件选择器。
- 引入 FILETREE 文件索引缓存。

## 任务拆分

### 后端

- 实现 `GET /api/v1/storage/drivers`。
- 实现 `GET /api/v1/storage/driverInfo`。
- 实现 `GET /api/v1/storage/files`。
- 实现 `GET /api/v1/storage/file`。
- 实现 `POST /api/v1/storage/files/remove`。
- 实现 `POST /api/v1/storage/file/rename`。
- 实现 `POST /api/v1/storage/files/copy`。
- 实现 `POST /api/v1/storage/files/move`。
- 请求 OpenList 时携带当前设备会话对应的 OpenList 认证信息。
- 将 OpenBridge 的 `/` 映射到当前 OpenList 用户可见根目录。
- 对 OpenList 源和用户根目录进行路径归一化处理。
- 删除、移动、复制、重命名后刷新 FILETREE 缓存。

### 缓存

- 维护 FILETREE 文件树缓存结构。
- 支持配置缓存磁盘上限，最小 4 KB。
- 支持配置缓存层数，范围 1 到 5。
- 启动时读取缓存，退出或重启时写回缓存。
- 缓存预热不得阻塞正常文件浏览。

### 前端

- OpenList 页面显示面包屑路径。
- 文件列表展示名称、大小、修改时间和操作。
- 支持按名称、大小、修改时间排序。
- 支持自动分页加载。
- 支持文件夹、图片、视频、音频、压缩包、PDF、Office、代码等类型图标。
- 支持表头复选框和行复选框。
- 支持工具栏复制、剪切、粘贴、删除、重命名、详细信息。
- 支持行内详情和重命名按钮。
- 支持 OpenList 路径选择器，用于下载任务源路径和其他路径输入场景。

## 验收标准

- 不同 OpenList 用户进入 `/openlist` 时看到自己的根目录。
- 文件超过 50 个时前端能继续加载。
- 删除、重命名、复制、移动操作能同步到 OpenList。
- 文件夹可进入，文件可查看详情和发起下载。
- 切换 OpenList Base URL 后文件浏览上下文不会串源。

## 交付物

- Storage API。
- OpenList 文件浏览器。
- OpenList 路径选择器。
- FILETREE 缓存能力。
