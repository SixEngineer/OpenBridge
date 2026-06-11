# Week 11 任务说明：最终验收、打包发布与链路闭环

## 本周定位

本周完成最终交付闭环。目标是让代码、文档、构建产物、Release、演示材料和端到端测试全部对应同一个版本。

## 本周目标

- 完成全链路回归测试。
- 完成前端生产构建并嵌入后端。
- 完成 Windows EXE 打包。
- 自动生成或提供 `.env.example`。
- 完成 Git 分支合并、版本标签和 GitHub Release。
- 上传发布附件。
- 完成用户手册、PDF、PPT 和演示视频交付。

## 任务拆分

### 测试闭环

- 测试登录、会话过期、后端重启、OpenList 换源和设备限制。
- 测试仪表盘资源指标和顶部栏带宽显示。
- 测试 Provider 新增、编辑、删除。
- 测试 Mount 创建、编辑、删除、配额同步和 WebDAV 容量展示。
- 测试 OpenList 文件浏览、分页、详情、删除、重命名、复制、剪切和粘贴。
- 测试单文件直链、二维码、当前设备下载和文件夹 ZIP。
- 测试百度直链清单。
- 测试 aria2 创建任务、进度、停止、批量停止、重试、打开文件、删除本地文件。
- 测试 rclone 配置写入、挂载、停止挂载和删除配置。
- 测试设置保存、服务重启、退出和自动打开浏览器。

### 构建发布

- 运行 `npm run build`。
- 将 `frontend/dist` 同步到 `backend/web/dist`。
- 运行 `go test ./...`。
- 运行 `go build -o openbridge.exe ./backend` 或等价构建命令。
- 确认 `APP_VERSION=v1.3`。
- 确认 `.env.example` 与实际配置项一致。
- 创建版本提交。
- 合并到 `dev`。
- 合并到 `main`。
- 创建并推送 `v1.3` 标签。
- 创建 GitHub Release。
- 上传 `openbridge.exe`、`.env.example` 和演示视频附件。

### 文档交付

- 用户手册 Markdown 与现有页面保持一致。
- 用户手册 PDF 可打开。
- API 文档覆盖当前接口。
- PPT 不超过 5 页，讲清功能概览、技术栈和软件规模。
- 演示视频和截图路径可追溯。
- 任务周计划整理为 Week 1 到 Week 11 单独文件。

## 验收标准

- `dev` 和 `main` 指向同一发布提交。
- `v1.3` 标签指向 `main` 发布提交。
- Release 页面存在且附件可下载。
- 打包 EXE 可启动服务并自动生成 `.env`。
- 浏览器能访问本地控制台。
- 主要功能链路均能完成一次端到端演示。
- 文档、PPT、视频和代码版本一致。

## 交付物

- `openbridge.exe`
- `.env.example`
- GitHub Release
- `v1.3` 标签
- 用户手册 Markdown 和 PDF
- 产品演示 PPT
- 产品演示视频
- Week 1 到 Week 11 任务说明文件
