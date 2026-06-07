# Week 2 - Backend Infrastructure Upgrade

## 本周目标

完成后端基础能力建设，使系统具备：

* 可配置（Config）
* 可观测（Logging + Request Trace）
* 可扩展的数据结构（DB Schema）

为后续 Provider、Quota、Task 系统打基础。

---

## 任务一：统一 API 响应结构（必须完成）

### 标准格式

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### 约定

| 字段      | 含义              |
| ------- | --------------- |
| code    | 0 表示成功，非 0 表示错误 |
| message | 简要说明            |
| data    | 返回数据            |

### 要求

* 所有 handler 必须统一使用该结构
* 禁止直接返回裸 JSON 或字符串
* 成功统一：`code = 0, message = "ok"`

---

## 任务二：错误体系（必须完成）

### 目标

建立统一 error code，避免混乱字符串错误

### 示例设计

```go
const (
    ErrCodeOK            = 0
    ErrCodeInvalidParam  = 1001
    ErrCodeUnauthorized  = 1002

    ErrCodeProviderFail  = 2001
    ErrCodeTokenExpired  = 2002

    ErrCodeInternalError = 9001
)
```

### 要求

* handler 不直接返回字符串错误
* 所有错误必须：

  * 有 code
  * 有 message
* usecase 返回 error 时必须可识别

---

## 任务三：配置系统（必须完成）

### `.env` 文件

```env
APP_NAME=OpenBridge
APP_ENV=dev
APP_PORT=8080

DB_PATH=./data/openbridge.db

ARIA2_RPC_URL=http://127.0.0.1:6800/jsonrpc
ARIA2_RPC_SECRET=

OPENLIST_BASE_URL=http://127.0.0.1:5244
OPENLIST_TOKEN=

LOG_LEVEL=debug
LOG_FORMAT=json
```

---

###  Config Struct

```go
type Config struct {
	App      AppConfig
	DB       DBConfig
	Aria2    Aria2Config
	OpenList OpenListConfig
	Log      LogConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DBConfig struct {
	Path string
}

type Aria2Config struct {
	RPCURL string
	Secret string
}

type OpenListConfig struct {
	BaseURL string
	Token   string
}

type LogConfig struct {
	Level  string
	Format string
}
```

---

### 要求

* 所有配置必须通过 Config 读取
* 禁止在代码中写死路径/URL
* 必须提供 `.env.example`
* 启动时做基础校验（DB_PATH 必须存在）

---

## 任务四：日志系统（建议完成）

### 推荐使用：zap

---

### 日志格式示例

```json
{
  "level": "info",
  "ts": "2026-04-07T10:20:00Z",
  "msg": "create access token",
  "request_id": "req_123456",
  "path": "/api/v1/token/access",
  "method": "POST",
  "status": 1000,
  "latency": 0.123,
  "user_id": "user_123456"
}
```

---

### 必须实现

#### request_id 中间件

功能：

* 每个请求生成唯一 request_id
* 写入：

  * gin context
  * response header

Header：

```
X-Request-ID
```

---

#### access log

记录：

* method
* path
* status
* latency
* request_id

---

### 日志分层建议

| 层级         | 内容      |
| ---------- | ------- |
| handler    | 请求入口    |
| usecase    | 业务逻辑    |
| repository | 仅 error |

---

##  任务五：数据库结构升级（必须完成）

---

### 5.1 Token 表

```go
type Token struct {
	ID           uint      `gorm:"primaryKey"`
	NetDisk      string    `gorm:"size:50;not null;index"`
	AccessToken  string    `gorm:"type:text;not null"`
	RefreshToken string    `gorm:"type:text"`
	ExpiresAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
```

---

### 设计说明

> 推荐：未来可以将 Token 直接合并进 `ProviderAccount`
> 当前阶段保留 Token 表，仅用于过渡

---

### 5.2 ProviderAccount 表（核心）

```go
type ProviderAccount struct {
	ID              uint      `gorm:"primaryKey"`
	Name            string    `gorm:"size:100;not null"`
	ProviderType    string    `gorm:"size:50;not null;index"`
	NetDisk         string    `gorm:"size:50;not null"`
	AccountID       string    `gorm:"size:100"`
	Status          string    `gorm:"size:20;not null;default:'active'"`
	AccessToken     string    `gorm:"type:text"`
	RefreshToken    string    `gorm:"type:text"`
	TokenExpiresAt  *time.Time
	TotalQuota      int64
	UsedQuota       int64
	AvailableQuota  int64
	LastQuotaSyncAt *time.Time
	LastError       string    `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
```

---

### 状态定义（建议）

```
active / disabled / expired / error
```

---

### 5.3 DownloadTask 表

```go
type DownloadTask struct {
	ID                uint      `gorm:"primaryKey"`
	TaskID            string    `gorm:"size:64;uniqueIndex;not null"`
	SourceURL         string    `gorm:"type:text;not null"`
	SourceType        string    `gorm:"size:50"`
	FileName          string    `gorm:"size:255"`
	FileSize          int64
	DirectLink        string    `gorm:"type:text"`
	ProviderAccountID *uint     `gorm:"index"`
	ProviderType      string    `gorm:"size:50;index"`
	Aria2GID          string    `gorm:"size:64;index"`
	Status            string    `gorm:"size:30;not null;index"`
	Progress          float64   `gorm:"default:0"`
	ErrorMessage      string    `gorm:"type:text"`
	RetryCount        int       `gorm:"default:0"`
	StartedAt         *time.Time
	FinishedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
```

---

### 状态建议（必须统一）

```
pending
resolving
resolved
submitting
downloading
completed
failed
cancelled
```

---

## 本周验收标准

* [ ] 所有 API 使用统一返回结构
* [ ] 错误码体系落地
* [ ] `.env + config` 可正常加载
* [ ] logger 初始化成功
* [ ] request_id 可贯穿请求
* [ ] DB 自动迁移成功
* [ ] 新三张表可写入数据

---



##  输出（代码结构）

```text
internal/
  config/
  logger/
  middleware/
  domain/entity/
```

---


