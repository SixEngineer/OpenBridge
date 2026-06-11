# Week 4 任务说明：Provider、Mount 与配额管理闭环

## 本周定位

本周完成服务商和 Mount 的核心模型，把 OpenBridge 从“能登录的控制台”推进到“能管理存储来源和配额”的系统。

## 本周目标

- 建立 Provider 抽象和服务商注册能力。
- 支持通用、百度、本地、夸克等 Provider 类型的统一管理。
- 建立 Mount 模型，绑定 Provider 与 OpenList 路径。
- 支持真实配额和虚拟配额。
- 前端完成服务商管理和配额管理页面。
- 服务商、Mount、配额按 OpenList Base URL 隔离。

## 任务拆分

### 后端

- 定义 ProviderAccount 实体。
- 实现 Provider Repository 和 UseCase。
- 实现 Provider 注册、列表、详情、编辑、删除接口。
- 定义 MountPoint 实体。
- 实现 Mount 创建、列表、编辑、删除接口。
- 定义配额模式：`real` 和 `virtual`。
- 实现 Mount 配额查询和同步。
- 记录 QuotaSnapshot。
- 对 Provider、Mount、QuotaSnapshot 加入 OpenList 源隔离字段。

### Provider 能力

- 通用 Provider：适配 OpenList 已有挂载路径。
- 百度 Provider：保存 access token，为后续直链解析准备。
- 本地 Provider：读取本机路径容量。
- 夸克 Provider：保存 Cookie 和账号标识，为后续容量同步准备。

### 前端

- 服务商管理页展示服务商卡片。
- 支持新增、编辑、删除服务商。
- 服务商表单支持不同 Provider 类型的字段切换。
- 配额管理页支持选择 Provider。
- 展示 Provider 总使用量和总容量。
- 支持创建、编辑和删除 Mount。
- 支持真实容量和虚拟容量显示。
- 修复下拉菜单遮挡问题。

## 验收标准

- 管理员可以新增、编辑、删除服务商。
- 管理员可以为服务商创建 Mount。
- Mount 能正确显示真实或虚拟配额。
- 普通用户只能查看服务商和基础信息，不能执行管理操作。
- 更换 OpenList 端口后不会看到旧源的 Provider 和 Mount。

## 交付物

- Provider API。
- Mount API。
- 配额同步 API。
- 服务商管理页面。
- 配额管理页面。
