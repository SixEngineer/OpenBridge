
API 接口文档：https://openbridge.apifox.cn

卢宇扬、郑源羽

#### 2026.04.06

按照项目规范搭建了后端的基本框架，详细如下。

```
backend/
├── main.go             # 程序入口文件
├── go.mod              
├── go.sum              
├── openbridge.db       # SQLite 数据库文件
├── internal/           # 内部业务代码
│   ├── domain/         # 领域层 - 定义核心业务模型和接口
│   ├── handler/        # 表现层 - HTTP 处理器，处理请求和响应
│   ├── repository/     # 数据访问层 - 数据库操作实现
│   ├── usecase/        # 应用层 - 业务逻辑实现
│   └── tool/           # 工具层 - 通用工具函数
└── pkg/                # 可被外部引用的公共包
    └── db/             # 数据库相关公共包
```

实现了以下两个提供给前端的接口，这里只做简要介绍，具体查看 API 接口文档。

- 上传AccessToken

由于大多数网盘都需要access token进行身份验证，因此需要将用户的access token上传到服务器，以便服务器可以代表用户进行操作。服务器将access token存储在数据库中，以便在需要时使用。

- 上传RefreshToken

由于access token通常具有较短的有效期，因此需要定期刷新access token。服务器将refresh token存储在数据库中，以便在需要时使用。

下一次开发任务：实现获取容量相关的接口。

#### 2026.04.16

- 实现了Zap日志系统，目前能够以JSON格式输出日志。

日志格式示例如下，这是一个HTTP请求的日志输出。

```json
{
    "level":"info",
    "ts":"2026-04-16T10:15:00.533+0800",
    "caller":"middleware/access_log.go:42",
    "msg":"",
    "request_id":"req_e98c1cbd7fb4978034497fff",
    "method":"GET",
    "path":"/api/v1/provider/info",
    "status":1000,
    "latency":0.0011538
}
```

- 实现了以下接口，具体信息查看 API 接口文档。

其中涵盖了provider的增删改查，quota的查询和同步。目前暂时使用mock数据作为返回内容。

POST   /api/v1/provider
DELETE /api/v1/provider
PUT    /api/v1/provider
GET    /api/v1/provider/info
GET    /api/v1/provider/list
POST   /api/v1/quota/query
POST   /api/v1/quota/sync

#### 2026.04.24

- 完成 Baidu Provider 首版接入，支持真实容量获取。

- 在保留现有 provider/quota 接口的前提下，新增 Mount + Quota Policy MVP，支持 `real / inherit / virtual` 三种模式及对应校验逻辑。

- 扩展 QuotaSnapshot，补充 mode、sync_status、error_message 等字段，用于记录每次配额解析结果与失败原因。

- 新增接口：

POST /api/v1/mount
GET  /api/v1/mount/:id/quota
POST /api/v1/mount/:id/quota/sync

- 已完成基础验证，原有 mock provider 与 quota query/sync 流程保持兼容。

#### 2026.04.30

TODO：完善 Refresh Token 机制，支持自动刷新 Access Token。

目前已经尝试由我们自己的 `refresh_token` 生成新的 `access_token`，但发现 Baidu 提供的 API 需要在百度开放平台配置一个应用，并申请上线审核，暂时无法实现。

#### 2026.06.07

- 修复 OpenFileLocation Windows 下路径含空格或相对路径时 explorer /select, 定位错误的问题
  - getActualFilePath 增加 filepath.Abs + strings.TrimSpace 确保返回绝对路径
  - OpenFile（cmd /c start）不受影响，因为 start 使用 ShellExecute Win32 API，可直接处理相对路径和含空格路径

- RetryTask: 修复重复 AddURI 的 bug，改为 if/else 分支
- GetTask: 同步 aria2 状态时处理 GID not found 情况，标记为 error
- 新增 DownloadUseCase.getActualFilePath 降级逻辑（DownloadDir + FileName 拼接） 