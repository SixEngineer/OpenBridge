# OpenBridge：面向 OpenList 的下载与挂载控制台

## 为什么做这个产品

- OpenList 能统一管理网盘文件，但“真正用起来”的最后一公里仍然分散：下载、挂载、容量展示、跨设备直链都要分别折腾。
- 手机端直链经常拿到 `127.0.0.1` 或 OpenList 代理地址，离开服务端电脑就不可用。
- rclone 命令配置复杂，多个 Mount 聚合、停挂、清理旧配置对普通用户不友好。
- Provider 容量、Mount 虚拟配额、aria2 下载任务、主机资源缺少一个统一控制台。

## 我们的目标

把 OpenList 从“文件入口”补成“可演示、可管理、可多设备使用”的本地网盘工作台。

![仪表盘总览](docs/screenshots/user_manual/09-dashboard-overview.png)

# 功能概览：按角色看系统能做什么

## 普通用户

- 登录后进入仪表盘，查看服务状态、主机资源、存储总量和带宽。
- 浏览当前 OpenList 用户根目录，支持排序、面包屑跳转、大目录完整加载。
- 单文件下载：当前设备下载、复制直链、二维码转移、提交到服务端 aria2。
- 文件夹下载：ZIP 打包下载，或对百度文件夹批量解析直链清单。
- 查看下载任务：筛选、排序、查看详情、复制直链、重试失败任务。

## 管理员

- 注册/编辑/删除 Provider：通用、百度、夸克、本地存储。
- 创建/编辑/删除 Mount，支持真实配额和虚拟配额。
- 配置 rclone：普通、union、combine，写入配置、挂载、停止挂载、删除配置。
- 配置 OpenList、aria2、rclone、本地路径选择、FILETREE 缓存、重启/退出服务。

![OpenList 文件浏览](docs/screenshots/user_manual/15-openlist-browser.png)

# 技术栈与系统架构

## 前端

- Vue 3 + TypeScript + Vite：组件化页面、类型约束、生产构建快。
- Pinia：集中管理登录态、Provider、Mount、任务 ID、用户角色和本地偏好。
- Vue Router：管理员与普通用户页面权限隔离。
- Axios：统一携带设备 ID，下载/大目录场景支持无限超时。

## 后端

- Go + Gin：轻量 HTTP API、静态前端内嵌、WebDAV 代理。
- GORM + SQLite：本地化部署，保存 Provider、Mount、任务、rclone 配置。
- OpenList API/WebDAV：复用 OpenList 文件能力和用户权限。
- aria2 JSON-RPC：服务端电脑执行下载。
- rclone CLI：自动写配置、挂载、停挂、清理。

## 架构流向

`浏览器/手机` → `OpenBridge 前端` → `Go API` → `OpenList / aria2 / rclone / 本地系统`

![下载确认弹窗](docs/screenshots/user_manual/19-download-dialog.png)

# 亮点与技术难点

## 亮点

- 多设备会话：设备 ID 绑定登录态，普通账号默认最多 5 台设备在线。
- 用户根目录隔离：不同 OpenList 用户看到的 `/` 对应各自可见根目录。
- 终端直下：百度等非 OpenList 代理直链可复制、二维码转移、当前设备直下。
- WebDAV Mount：文件能力复用 OpenList，容量展示改写为 Mount 层配额。
- rclone 管理闭环：从配置、写入、挂载到停挂/删除，全流程前端化。

## 难点

- 大文件夹扫描与下载不能被 30 秒超时打断，需要前后端下载链路无限超时。
- FILETREE 缓存要加速文件树，但不能阻塞正常文件浏览。
- OpenList 换源后 Provider、Mount、rclone 配置必须按端口/源隔离。
- Windows 下打开文件夹、中文文件名、特殊符号路径要避免 shell 转义坑。
- WebDAV `PROPFIND` 响应要改写 `href`、`quota-used-bytes`、`quota-available-bytes`。

![Rclone 配置](docs/screenshots/user_manual/54-rclone-card.png)

# 软件规模与 10 分钟展示建议

## 软件规模

- 功能用例总数：42 个，按用户故事合并统计。
- 代码源文件总数：117 个，统计后端 Go + 前端 Vue/TS/CSS/HTML/JSON。
- 代码总行数：21459 行，排除 `node_modules`、`dist`、截图和文档。
- 后端 API/路由：49 个。
- 前端路由：11 个。

## 规模背后的模块

- 前端页面：仪表盘、OpenList、服务商、下载任务、配额、Rclone、设置、Debug。
- 后端用例层：用户、Provider、Mount、Storage、Download、Rclone、Settings、System。
- 外部集成：OpenList、aria2、rclone、WebDAV、本地文件选择和主机指标。

## 推荐节奏

- 0:00-1:30：为什么做，讲 OpenList 的“最后一公里”痛点。
- 1:30-3:20：功能概览，按普通用户和管理员两条线讲。
- 3:20-5:20：技术栈与难点，突出 Go + Vue + OpenList/aria2/rclone/WebDAV 集成。
- 5:20-9:20：演示 4 个闭环：文件浏览 → 下载弹窗 → 配额/Mount → rclone 配置。
- 9:20-10:00：总结亮点：多设备、直链、Mount 配额、rclone 闭环、大目录无限超时。

![下载任务](docs/screenshots/user_manual/26-tasks-overview.png)
