# OpenBridge 前端开发日志

> 说明：本文档不是最终方案文档，也不是结题总结。
> 它的用途是记录前端到当前为止做了什么、为什么这样做、现在处于什么状态，后续继续开发时应持续追加，不断完善。

---

## 日志使用规则

- 这份文档按“开发日志”维护
- 每完成一项前端工作，就在文档中新增记录
- 记录重点是：做了什么、当前状态、涉及文件、后续待做
- 不写最终结论，不写终局判断
- 文档应随着项目推进持续补充

---

## 2026-04-07 前端方向确认

### 本次确认内容

- 明确项目文档中前端技术栈为 `Vue3`
- 明确当前仓库里原始的 [frontend](/e:/se_work/Monorepo/frontend) 目录尚未正式开发
- 明确当前后端代码实现还比较早，前端可以先独立推进
- 明确前端现阶段应以 mock 数据驱动开发，而不是等待完整后端

### 当前理解

根据当前仓库与文档，`OpenBridge` 前端更适合做成一个“管理控制台”，而不是普通网盘页面或纯展示页。

页面重点应围绕以下模块展开：

- Dashboard
- OpenList
- Providers
- Download Tasks
- Quota
- Settings
- Debug

### 当前状态

- 前端可以启动独立搭建
- 不必等待后端全部完成
- 当前开发应以“先搭控制台骨架”为主

---

## 2026-04-07 页面结构规划整理

### 本次完成内容

根据项目定位，整理出一套适合 `OpenBridge` 的前端页面结构。

### 规划结果

建议前端页面结构为：

```text
OpenBridge Frontend
├─ 登录 / 入口
├─ Dashboard
├─ OpenList
├─ Providers
├─ Download Tasks
├─ Quota
├─ Settings
└─ Debug
```

### 页面作用说明

#### Dashboard

- 展示系统总览
- 展示运行状态
- 展示任务、告警、服务健康情况

#### OpenList

- 展示 OpenList 接入相关功能入口
- 后续可放连接测试、驱动列表、目录浏览等内容

#### Providers

- 展示 Provider 列表和状态
- 展示能力信息和认证方式

#### Download Tasks

- 展示下载任务列表
- 展示进度、状态、目标路径等内容

#### Quota

- 展示容量信息
- 展示各 Provider 的使用情况

#### Settings

- 放置配置相关页面

#### Debug

- 放置调试、排错和链路分析内容

### 当前状态

- 页面结构已形成统一认识
- 后续开发可以按这套结构推进

---

## 2026-04-07 控制台原型建立

### 本次完成内容

独立创建了一个新的前端原型目录：

- [frontend_test](/e:/se_work/Monorepo/frontend_test)

该目录不影响原始 [frontend](/e:/se_work/Monorepo/frontend)，作为单独的前端原型开发空间。

### 当前技术实现

在 `frontend_test` 中已完成以下基础搭建：

- `Vue3`
- `Vite`
- `TypeScript`
- `Vue Router`
- `Pinia`

### 当前已建立的项目结构

已建立的主要结构包括：

- 路由入口
- 全局入口
- 页面目录
- 组件目录
- 样式目录
- mock 数据目录
- store 目录

### 当前已建立的页面

当前已经建立这些页面文件：

- [DashboardView.vue](/e:/se_work/Monorepo/frontend_test/src/views/DashboardView.vue)
- [OpenListView.vue](/e:/se_work/Monorepo/frontend_test/src/views/OpenListView.vue)
- [ProviderView.vue](/e:/se_work/Monorepo/frontend_test/src/views/ProviderView.vue)
- [DownloadTasksView.vue](/e:/se_work/Monorepo/frontend_test/src/views/DownloadTasksView.vue)
- [QuotaView.vue](/e:/se_work/Monorepo/frontend_test/src/views/QuotaView.vue)
- [SettingsView.vue](/e:/se_work/Monorepo/frontend_test/src/views/SettingsView.vue)
- [DebugView.vue](/e:/se_work/Monorepo/frontend_test/src/views/DebugView.vue)
- [LoginView.vue](/e:/se_work/Monorepo/frontend_test/src/views/LoginView.vue)

### 当前已建立的布局组件

已建立的主要布局文件：

- [AppShell.vue](/e:/se_work/Monorepo/frontend_test/src/components/layout/AppShell.vue)
- [AppSidebar.vue](/e:/se_work/Monorepo/frontend_test/src/components/layout/AppSidebar.vue)
- [AppTopbar.vue](/e:/se_work/Monorepo/frontend_test/src/components/layout/AppTopbar.vue)

### 当前已建立的基础通用组件

- [PageHeader.vue](/e:/se_work/Monorepo/frontend_test/src/components/common/PageHeader.vue)
- [MetricCard.vue](/e:/se_work/Monorepo/frontend_test/src/components/common/MetricCard.vue)
- [StatusBadge.vue](/e:/se_work/Monorepo/frontend_test/src/components/common/StatusBadge.vue)

### 当前样式情况

已建立全局样式文件：

- [index.css](/e:/se_work/Monorepo/frontend_test/src/styles/index.css)

当前样式方向是：

- 控制台式布局
- 左侧导航 + 顶部栏 + 主内容区
- 玻璃感卡片
- 偏工程系统风格

### 当前状态

- 控制台壳子已经建立
- 路由与页面关系已经打通
- 当前页面已有基础视觉效果
- 多数业务页目前仍属于“结构占位 + 部分演示数据”阶段

---

## 2026-04-07 mock 数据接入

### 本次完成内容

为避免等待后端接口，已使用 mock 数据驱动控制台原型。

### 当前 mock 数据文件

- [dashboard.ts](/e:/se_work/Monorepo/frontend_test/src/mock/dashboard.ts)
- [provider.ts](/e:/se_work/Monorepo/frontend_test/src/mock/provider.ts)
- [tasks.ts](/e:/se_work/Monorepo/frontend_test/src/mock/tasks.ts)
- [quota.ts](/e:/se_work/Monorepo/frontend_test/src/mock/quota.ts)

### 当前 store 文件

- [console.ts](/e:/se_work/Monorepo/frontend_test/src/stores/console.ts)

### 当前 mock 数据作用

当前这些数据用于支撑：

- Dashboard 指标卡片
- 服务健康状态
- 最近任务
- 告警信息
- Provider 列表
- 下载任务列表
- Quota 卡片

### 当前状态

- 前端展示内容不依赖真实接口
- 可以先独立继续做页面
- 后续联调时需要逐步替换为真实接口数据

---

## 2026-04-07 关于登录页面的处理说明

### 当前情况

在 `frontend_test` 中已建立：

- [LoginView.vue](/e:/se_work/Monorepo/frontend_test/src/views/LoginView.vue)

但该页面当前只是静态页面占位，不是正式登录流程。

### 当前未实现内容

以下内容目前都还没有做：

- token 保存
- 登录状态管理
- 路由守卫
- 权限控制
- 登出逻辑

### 当前状态说明

也就是说，当前 `frontend_test` 更接近“控制台原型”而不是“完整带鉴权系统的前端”。

这个问题在讨论中已经被指出，后续不应把现在这套登录占位理解为最终方案。

---

## 2026-04-07 免登录门户方案确认

### 本次确认内容

在讨论中提出了一种新的入口方式：

- 不使用传统登录页作为主要入口
- 使用一个视觉上更有设计感的 `OpenBridge` 字标启动页
- 只要 `OpenList` 没断开，就允许进入控制台

### 当前方案理解

该方案的核心逻辑是：

1. 用户先看到一个沉浸式 `OpenBridge` 门户页
2. 门户页展示当前 OpenList 连接状态
3. 若连接正常，点击字标即可进入控制台
4. 控制台本体保持独立

### 当前状态

- 已形成明确方案
- 方案偏向课程项目展示效果
- 暂不走传统登录页路线

---

## 2026-04-07 免登录门户原型建立

### 本次完成内容

独立创建了一个新的门户目录：

- [frontend_portal](/e:/se_work/Monorepo/frontend_portal)

该目录同样不影响原始 [frontend](/e:/se_work/Monorepo/frontend) 和 [frontend_test](/e:/se_work/Monorepo/frontend_test)。

### 当前已建立内容

已建立以下文件：

- [package.json](/e:/se_work/Monorepo/frontend_portal/package.json)
- [main.ts](/e:/se_work/Monorepo/frontend_portal/src/main.ts)
- [App.vue](/e:/se_work/Monorepo/frontend_portal/src/App.vue)
- [styles.css](/e:/se_work/Monorepo/frontend_portal/src/styles.css)

### 当前门户页功能

当前门户页实现了：

- 大型 `OpenBridge` 字标展示
- OpenList 连接状态显示
- 点击字标进入控制台
- 点击按钮进入控制台

### 当前跳转目标

当前门户页目标地址为：

```text
http://localhost:5173/dashboard
```

即指向 `frontend_test` 控制台。

### 当前状态

- 免登录门户原型已建立
- 门户页和控制台已分离
- 当前连接状态还是演示逻辑，未接真实接口

---

## 2026-04-08 门户页视觉精简与说明收敛

### 本次完成内容

对 `frontend_portal` 门户页做了一轮收敛式调整，目标是把页面从“演示说明页”继续压缩成更纯粹的“启动入口页”。

本次主要完成：

- 移除右上角的演示用连接状态切换开关
- 移除底部 3 个解释性信息卡片
- 移除按钮下方的目标地址说明文案，不再在页面上暴露跳转目标
- 调整 `OpenBridge` 字标样式，缓解 `g` 下沿和 `e` 右侧的视觉截断感
- 保留左上角连接状态和中间主按钮，继续维持“单页入口”结构

### 涉及文件

- [App.vue](/e:/se_work/Monorepo/frontend_portal/src/App.vue)
- [styles.css](/e:/se_work/Monorepo/frontend_portal/src/styles.css)

### 调整原因

- 右上角开关会让页面过于像演示面板，而不是正式入口
- 底部说明卡片信息密度低，会稀释入口页的焦点
- 跳转目标地址属于开发实现细节，不适合直接展示给用户
- 门户页目前的核心职责是“确认可进入并发起进入”，不需要承担解释过多背景信息的任务

### 当前状态

- 门户页结构已经明显简化
- 页面视觉焦点集中在品牌字标、连接状态和进入动作
- 当前仍保留本地演示逻辑：连接状态为前端内部状态，点击后跳转到 `http://localhost:5173/dashboard`
- `frontend_portal` 现阶段更适合被理解为“控制台启动页原型”，而不是完整业务首页

### 后续待做

- 将连接状态从本地 `ref` 状态切换为真实健康检查结果
- 决定门户页是否保留英文文案，或统一改为中文
- 如果后续继续优化视觉，可考虑将 `OpenBridge` 字标替换为更稳定的品牌化方案，而不是仅靠超粗文字样式支撑

---

## 2026-04-08 正式前端目录收敛

### 本次完成内容

将此前分散在 `frontend_test` 与 `frontend_portal` 两个原型目录中的成果正式收敛到 [frontend](/e:/se_work/Monorepo/frontend)。

本次主要完成：

- 以 `frontend_test` 作为正式前端工程底座，补全 `frontend` 目录的 Vue + Vite + TypeScript 工程结构
- 将门户页整合进正式前端，新增首页入口页 [PortalView.vue](/e:/se_work/Monorepo/frontend/src/views/PortalView.vue)
- 调整路由结构，使 `/` 作为门户页，`/dashboard` 作为控制台首页
- 将门户页点击进入的行为从“跳到外部原型地址”改为“跳到同一前端工程内的控制台路由”
- 为正式前端目录安装依赖并完成构建验证
- 清理原型目录 [frontend_test](/e:/se_work/Monorepo/frontend_test) 与 [frontend_portal](/e:/se_work/Monorepo/frontend_portal)，避免团队后续继续基于临时目录开发

### 涉及文件

- [package.json](/e:/se_work/Monorepo/frontend/package.json)
- [vite.config.ts](/e:/se_work/Monorepo/frontend/vite.config.ts)
- [index.ts](/e:/se_work/Monorepo/frontend/src/router/index.ts)
- [PortalView.vue](/e:/se_work/Monorepo/frontend/src/views/PortalView.vue)
- [frontend_开发文档.md](/e:/se_work/frontend_开发文档.md)

### 调整原因

- 团队协作时应只保留一个正式前端目录，避免多人分别在不同原型目录继续开发
- `frontend_test` 与 `frontend_portal` 各自承担了一部分能力，但并行存在不利于分支协作、联调和答辩说明
- 将门户页与控制台合并到同一工程后，项目结构更符合 Monorepo 下正式应用目录的定位

### 当前状态

- [frontend](/e:/se_work/Monorepo/frontend) 已成为唯一正式前端工程
- 正式前端已可通过 `npm run dev` 启动，并通过 `npm run build` 构建验证
- 首页为门户页，门户页进入后跳转到同工程内的 `/dashboard`
- 旧原型目录已完成阶段使命，不再作为后续开发目标

### 后续待做

- 将控制台页面中的演示数据逐步替换为真实接口
- 视团队决定，统一前端页面文案语言风格
- 后续所有前端新功能都应基于 [frontend](/e:/se_work/Monorepo/frontend) 继续推进

---

## 当前目录说明

### 原始目录

- [frontend](/e:/se_work/Monorepo/frontend)
  - 当前为正式前端工程，包含门户页与控制台

### 当前目录状态

- `frontend_test`
  - 原控制台原型目录，现已并入 `frontend`

- `frontend_portal`
  - 原门户页原型目录，现已并入 `frontend`

### 当前文档文件

- [frontend_开发文档.md](/e:/se_work/Monorepo/frontend_开发文档.md)
  - 用途：持续记录前端开发进度

---

## 当前已知待继续完善内容

以下内容目前还没有完成，后续开发时可继续往日志中追加：

- `frontend` 页面中文化
- Dashboard 内容继续细化
- OpenList 页面从占位升级为真实结构
- Providers 页面增加更多管理交互
- Tasks 页面增加详情、筛选、状态动作
- Quota 页面增加更清晰的数据展示
- Settings 页面补充真实配置表单
- Debug 页面补充调试结构
- mock 数据逐步向真实接口字段收拢
- 门户页连接状态改为真实健康检查

---

## 2026-04-19 前后端接口对接

### 本次完成内容

完成前端控制台核心业务模块与后端接口的对接，将此前基于 mock 数据的页面替换为真实 API 调用。

本次主要完成：

- **Provider 模块对接**
  - 实现 Provider 列表获取
  - 实现 Provider 注册功能
  - 实现 Provider 编辑功能
  - 实现 Provider 删除功能
  - 新增 ProviderFormDialog 表单对话框组件

- **Quota 模块对接**
  - 实现配额查询功能（不触发远端同步）
  - 实现配额同步功能（调用远端接口）
  - 展示总配额、已用配额、可用配额及使用进度条

- **Token 模块对接**
  - 新增 Token 管理页面
  - 实现 Token 上传功能
  - 支持选择网盘类型（mock/baidu/aliyun/quark）

- **基础设施完善**
  - 安装并配置 axios 请求库
  - 创建统一的请求封装 `utils/request.ts`
  - 配置 Vite 代理解决跨域问题
  - 完善 TypeScript 类型定义，匹配后端数据结构
  - 调整状态管理 `stores/console.ts`，接入真实 API
  - 适配前端状态显示组件（StatusBadge）以匹配后端状态值（active/disabled/expired/error）

### 涉及文件

**新增文件：**
- [frontend/src/utils/request.ts](/e:/se_work/Monorepo/frontend/src/utils/request.ts)
- [frontend/src/api/provider.ts](/e:/se_work/Monorepo/frontend/src/api/provider.ts)
- [frontend/src/api/quota.ts](/e:/se_work/Monorepo/frontend/src/api/quota.ts)
- [frontend/src/api/token.ts](/e:/se_work/Monorepo/frontend/src/api/token.ts)
- [frontend/src/types/quota.ts](/e:/se_work/Monorepo/frontend/src/types/quota.ts)
- [frontend/src/types/token.ts](/e:/se_work/Monorepo/frontend/src/types/token.ts)
- [frontend/src/views/TokenView.vue](/e:/se_work/Monorepo/frontend/src/views/TokenView.vue)
- [frontend/src/components/provider/ProviderFormDialog.vue](/e:/se_work/Monorepo/frontend/src/components/provider/ProviderFormDialog.vue)
- [frontend/.env.development](/e:/se_work/Monorepo/frontend/.env.development)

**修改文件：**
- [frontend/package.json](/e:/se_work/Monorepo/frontend/package.json)（添加 axios 依赖）
- [frontend/vite.config.ts](/e:/se_work/Monorepo/frontend/vite.config.ts)（添加代理配置）
- [frontend/src/stores/console.ts](/e:/se_work/Monorepo/frontend/src/stores/console.ts)
- [frontend/src/views/ProviderView.vue](/e:/se_work/Monorepo/frontend/src/views/ProviderView.vue)
- [frontend/src/views/QuotaView.vue](/e:/se_work/Monorepo/frontend/src/views/QuotaView.vue)
- [frontend/src/types/provider.ts](/e:/se_work/Monorepo/frontend/src/types/provider.ts)
- [frontend/src/types/common.ts](/e:/se_work/Monorepo/frontend/src/types/common.ts)
- [frontend/src/components/common/PageHeader.vue](/e:/se_work/Monorepo/frontend/src/components/common/PageHeader.vue)
- [frontend/src/components/common/StatusBadge.vue](/e:/se_work/Monorepo/frontend/src/components/common/StatusBadge.vue)
- [frontend/src/components/layout/AppSidebar.vue](/e:/se_work/Monorepo/frontend/src/components/layout/AppSidebar.vue)
- [frontend/src/router/index.ts](/e:/se_work/Monorepo/frontend/src/router/index.ts)
- [frontend/src/styles/index.css](/e:/se_work/Monorepo/frontend/src/styles/index.css)

### 当前状态

- Provider 模块已完成完整的增删改查功能，数据来源于真实后端接口
- Quota 模块已完成查询与同步功能，可正确展示配额数据
- Token 模块已完成上传功能，可向指定网盘类型上传 Token
- 前端通过 Vite 代理方式解决跨域问题，无需依赖后端 CORS 配置
- 所有对接模块均已在本地完成测试，功能正常运行

### 后续待做

- Dashboard 页面数据接入真实接口
- OpenList 页面从占位升级为真实功能
- Download Tasks 页面等待后端接口完成后进行对接
- Quota 历史快照展示（等待后端暴露 QuotaSnapshot 接口）
- 统一前后端响应码处理逻辑
- 完善全局错误处理和用户提示

---

## 2026-06-05 前端交互细节优化

### 本次完成内容

对前端多个页面进行了交互细节和 UI 优化，包括复选框多选、排序、配额展示、管理员模式等。

- **下载任务页复选框多选**
  - 新增列首全选复选框（始终可见，浅蓝色 `accent-color: #bfdbfe`）
  - 行复选框默认不可见，hover 行时显示
  - 选中后复选框保持可见
  - 支持全选/取消全选，切换筛选标签时清空选中
  - 统一清除按钮：有选中项时显示”清除 (N)”，有已完成任务时显示”清空全部”

- **移除刷新按钮**
  - 下载任务页移除手动刷新按钮，刷新浏览器即可

- **OpenList 默认排序改为降序**
  - `sortOrder` 默认值从 `'asc'` 改为 `'desc'`

- **配额管理页简化**
  - 移除”查询配额”按钮，只保留”同步配额”
  - 进入页面后自动同步一次配额

- **Provider 表单调整**
  - 注册按钮与描述文字对齐

- **页面头部调整**
  - 移除所有页面左上角的”模块”小字

- **登录/退出流程修复**
  - 退出登录后跳转到 `/login` 而非首页
  - 修复退出后仍能查看数据的问题

- **管理员模式**
  - 双击顶部栏 OpenList 状态指示器切换管理员模式
  - 管理员模式持久化到 localStorage
  - 侧边栏 Debug 页面仅管理员可见
  - 路由守卫拦截非管理员访问 `/debug`

- **Mock Provider 配额单位修正**
  - Mock 类型配额从 MB 改为 GB 展示（Dashboard/Provider/Quota 三个页面）
  - Mock 数据源数值重新计算（1000 MB → 1024000 GB 等）

- **Dashboard 存储用量卡片交互**
  - 点击”存储用量”指标卡展开所有 Provider 配额明细
  - 展开后可点击选择某一个作为默认展示（主指标卡显示选中 Provider 的名称和数据）

- **i18n 翻译修正**
  - “Provider” 统一翻译为”服务商”
  - 新增 `tasks.clear_selected` 等翻译键

### 涉及文件

- `frontend/src/views/DownloadTasksView.vue`
- `frontend/src/views/DashboardView.vue`
- `frontend/src/views/QuotaView.vue`
- `frontend/src/views/ProviderView.vue`
- `frontend/src/views/OpenListView.vue`
- `frontend/src/components/common/PageHeader.vue`
- `frontend/src/components/common/MetricCard.vue`
- `frontend/src/components/layout/AppTopbar.vue`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/stores/console.ts`
- `frontend/src/router/index.ts`
- `frontend/src/i18n/locales/zh-CN.json`
- `frontend/src/i18n/locales/en.json`
- `frontend/src/mock/provider.ts`

### 当前状态

- 下载任务页支持复选框多选和批量清除
- Dashboard 存储用量卡片支持展开和选择默认 Provider
- 管理员模式已实现，可以保护敏感页面
- Mock 数据配额单位已对齐 MB/GB
- 页面整体交互流畅度提升

### 后续待做

- 下载页复选框样式微调（对齐、选中态可见性）

---

## 2026-06-06 前端细节修复与 Provider 表单增强

### 本次完成内容

修复了多项 UI 细节问题，增强 Provider 注册表单的易用性，完善国际化文案。

- **复选框交互修复**
  - 选中（checked）的复选框保持可见，不依赖 hover 状态
  - 移除 `padding-top: 2px` 使复选框与文件名文字对齐

- **Settings 页面 OpenList 卡片精简**
  - 移除状态字段中的驱动数量文字（仍保留连接状态圆点指示器）
  - 新增 `openlist_connected` 翻译键

- **侧边栏翻译优化**
  - `openlist_desc`: “网盘源访问” → “文件浏览”
  - `providers_desc`: “适配器注册” → “服务商管理”

- **Provider 注册表单增强**
  - 百度网盘：access_token 输入框旁新增”点我获取 Token”按钮，跳转 https://api.oplist.org
  - 夸克网盘：Cookie 输入框旁新增”点我获取 Cookie”按钮，跳转 https://pan.quark.cn
  - 百度提示文案改为”通过百度网盘验证登录填写对应的 Key，获取令牌”
  - 后端类型选项”通用（推荐）”改为”通用”

- **下载任务详情增强**
  - 直链行新增”点击复制”按钮（复制到剪贴板，短暂显示”已复制!”）
  - 直链改为横向滚动显示，不再撑破容器

- **OpenList 文件浏览器修复**
  - 过滤 API 返回的 `provider: “unknown”`，改为显示当前路径

- **国际化同步**
  - 同步更新 en.json 中对应的翻译文案（Baidu hint、Generic、copy_link、get_cookie 等）
  - 新增 `tasks.copy_link`、`provider_form.get_token`、`provider_form.get_cookie` 翻译键

### 涉及文件

- `frontend/src/views/DownloadTasksView.vue`
- `frontend/src/views/OpenListView.vue`
- `frontend/src/views/SettingsView.vue`
- `frontend/src/components/provider/ProviderFormDialog.vue`
- `frontend/src/i18n/locales/zh-CN.json`
- `frontend/src/i18n/locales/en.json`
- `docs/frontend_dev.md`

### 当前状态

- Provider 表单对百度网盘和夸克网盘提供了快捷跳转入口，降低用户操作成本
- 直链复制功能已可用
- 翻译文案持续优化，中英文同步维护
- 复选框交互符合预期：未选中的 hover 显示，选中的常驻可见

### 后续待做

- 下载任务详情面板响应式布局优化
- Provider 表单校验增强（如 token/cookie 非空校验）
- 下载任务页选中项批量操作功能完善（如批量重新下载）

## 2026-05-28 前端交互链路完善

### 本次完成内容

完善了前端多个模块的交互链路，打通从 Provider 注册 → Mount 创建 → 配额查看 → 文件浏览 → 下载的完整流程。

- **Provider 表单重构**
  - 将表单从"选择 OpenList 驱动"改为直接选择后端支持的三种 Provider 类型（通用/百度网盘/本地存储）
  - 百度网盘类型显示 access_token 输入框
  - 本地存储类型显示路径输入框，路径存入 account_id 字段
  - Provider 卡片页对本地类型显示"本地路径"而非"账户ID"

- **Quota 页面完善**
  - 修复 MB 显示问题（后端以 MB 为单位返回，前端 formatBytes 按字节处理），增加 formatQuotaMB 中间转换函数
  - 创建 Mount 时支持选择配额模式：Real（真实容量）/ Virtual（虚拟容量），Inherit 预留
  - Virtual 模式可输入虚拟总容量（MB）
  - 增加操作状态反馈提示（成功/失败）

- **OpenList 文件列表增加下载入口**
  - 文件列表每行增加"下载"按钮（仅文件类型）
  - 点击后弹出下载确认弹窗
  - 弹窗自动解析直链，显示文件名/大小/Provider/直链/是否走代理
  - 可选填下载目录，确认后提交到 aria2

- **下载任务列表改造**
  - 从单任务创建页改为任务列表视图
  - 显示所有历史任务，支持状态筛选（All/Active/Completed/Failed等）
  - 每个任务显示文件名/大小/状态/进度/创建时间
  - 点击行展开详情面板
  - 活跃任务自动每 5 秒刷新，完成后自动停止

### 涉及文件

**新增文件：**
- `frontend/src/components/download/DownloadDialog.vue`

**修改文件：**
- `frontend/src/views/OpenListView.vue`
- `frontend/src/views/QuotaView.vue`
- `frontend/src/views/DownloadTasksView.vue`
- `frontend/src/views/ProviderView.vue`
- `frontend/src/components/provider/ProviderFormDialog.vue`
- `frontend/src/stores/console.ts`
- `frontend/src/api/task.ts`
- `frontend/src/types/download.ts`

### 当前状态

- Provider 表单已覆盖后端支持的三种类型
- Quota 页面已支持 Real/Virtual 配额模式，Mount 创建后可查询和同步
- 文件浏览到下载的链路已打通：OpenList 文件列表 → 下载弹窗 → 确认 → aria2
- 下载任务列表可查看所有历史任务和实时进度
- 本次所有改动不涉及后端代码

### 后续待做

- 下载确认弹窗成功后跳转到任务列表的交互优化
- Inherit 配额模式需后端暴露 mount 列表 API 后再启用
- aria2 未启动时的错误提示优化
- 任务列表手动刷新按钮的加载状态细化

---

## 2026-06-06 Settings/Debug 页面重构与路由权限调整

### 本次完成内容

对 Settings、Dashboard、Debug 三个页面进行了功能重分布，并调整了 Debug 页面的访问权限策略。

- **Settings 页面精简**
  - 移除只读状态信息（API 连接、OpenList 状态、服务商统计）
  - 仅保留用户可配置项：aria2 RPC URL（管理员可修改）和默认下载目录
  - OpenList 基础 URL 字段，仅管理员可修改
  - 已连接状态用圆点指示器简化显示

- **Dashboard 页面调整**
  - OpenList 状态卡片替换为 aria2 RPC 连接状态
  - 系统健康检测增加 aria2 状态检查（通过直接 JSON-RPC 调用 `aria2.getVersion`）
  - 保留顶部 OpenList 连接状态指示器

- **Debug 页面完善**
  - 新增 API 连接信息面板（Base URL / Proxy Target / Timeout）
  - 重置用户数据功能从 Settings 移至 Debug
  - 重置按钮仅管理员可见

- **权限策略调整**
  - Debug 页面改为所有用户可见（移除路由 `meta.admin` 守卫和侧边栏过滤）
  - 只有重置数据按钮受管理员权限控制

- **其他修复**
  - Provider 表单百度/夸克增加外部链接按钮
  - 下载任务详情直链增加复制按钮
  - OpenList 文件浏览器过滤 `provider: "unknown"`
  - 复选框对齐修复

### 涉及文件

- `frontend/src/views/DashboardView.vue`
- `frontend/src/views/SettingsView.vue`
- `frontend/src/views/DebugView.vue`
- `frontend/src/router/index.ts`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/components/provider/ProviderFormDialog.vue`
- `frontend/src/views/OpenListView.vue`
- `frontend/src/views/DownloadTasksView.vue`
- `frontend/src/i18n/locales/zh-CN.json`
- `frontend/src/i18n/locales/en.json`
- `docs/frontend_dev.md`

### 当前状态

- Settings 页面功能定位明确：仅展示用户可配置项
- Dashboard 更加专注系统实时状态（aria2 RPC 取代 OpenList 状态卡）
- Debug 作为诊断入口，对所有用户可见，破坏性操作受管理员权限保护
- 管理员模式已从"页面级拦截"改为"操作级控制"

---

## 2026-06-06 下载状态重构与交互完善

### 本次完成内容

对下载任务的状态机逻辑进行全面重构，联动后端修复 aria2 离线检测，优化轮询策略、清空按钮行为、登录键盘导航和 Dashboard 交互细节。

- **下载状态机重构（前后端协作）**
  - 后端 `CreateTask`：初始状态从 `"active"` 改为 `"waiting"`，更准确表达刚提交尚未开始的状态
  - 后端 `GetTask`：TellStatus 成功时写入 `FinishedAt` 时间戳；TellStatus 返回"GID not found"且当前不是 `complete` 时将任务标记为 `error`
  - 修复 aria2 错误消息匹配：aria2 实际返回 `"GID xxxx is not found"`（非文档中的 `"No such download"`），字符串匹配改为 `strings.Contains(err.Error(), "not found")`
  - 修复全部标记为 error 的问题：增加 `&& task.Status != "complete"` 保护，已完成的任务不受影响
  - aria2 网络错误（如连接被拒）保留当前状态，不误判为 error
  - 预留空 GID 检查和类型断言安全保护

- **轮询优化（前端）**
  - 移除 12 次硬限制（原 12×25s=5min），改为无限制轮询
  - 间隔从 25s 缩短到 5s
  - 仅当存在活跃任务（等待/下载中）时保持轮询，全部结束后自动停止

- **清空按钮行为重写**
  - 根据当前筛选标签删除任务（如"失败"标签只删失败任务）
  - 不再全部页面一刀切删除
  - 未选中时显示"清空当前"，有选中项时显示"清除 (N)"

- **Dashboard 挂载点卡片改进**
  - 移除活跃挂载点下方的已用量显示（无实际意义）
  - 实现点开展示所有挂载点列表
  - 展开挂载点时自动折叠存储用量展开，反之亦然（互斥展开）

- **登录页键盘导航**
  - 进入页面自动聚焦用户名输入框（`autofocus`）
  - 用户名框按回车跳转到密码框 `passwordInput.value?.focus()`
  - 密码框按回车触发登录
  - 用户名已填但密码为空时聚焦密码框而不是显示错误提示

- **MetricCard 组件优化**
  - trend 字段为空时不渲染空白行（`v-if="item.trend"`）

- **下载任务详情增强**
  - 直链行新增"点击复制"按钮（Clipboard API）
  - 直链横向滚动显示
  - 已完成界面显示完成时间（`FinishedAt`），其他标签显示创建时间

- **国际化补充**
  - zh-CN：`finished_col`、"完成时间"、`clear_filtered`、"清空当前"
  - en.json：`finished_col`、"Finished"、`clear_filtered`、"Clear Current"

### 涉及文件

- `frontend/src/views/DownloadTasksView.vue`（状态映射、轮询、清空、详情面板）
- `frontend/src/views/DashboardView.vue`（挂载点展开、互斥逻辑）
- `frontend/src/views/LoginView.vue`（键盘导航、自动聚焦）
- `frontend/src/components/common/MetricCard.vue`（条件渲染 trend）
- `frontend/src/i18n/locales/zh-CN.json`
- `frontend/src/i18n/locales/en.json`
- `backend/internal/usecase/download_usecase.go`（状态机、FinishedAt、error 检测）

### 当前状态

- 下载状态机逻辑完整：waiting → active → complete/error，aria2 离线可正确检测
- 轮询 5s 间隔实时感知任务状态变化，无活跃任务时零开销
- 清空按钮行为匹配用户当前筛选视图，不再误删
- Dashboard 和登录页交互更流畅

### 后续待做

- 批量下载操作（全选后批量重试/删除）
- aria2 RPC 重连机制（aria2 重启后自动恢复任务状态同步）
- 错误详情弹窗（点击失败任务查看具体错误信息）

---