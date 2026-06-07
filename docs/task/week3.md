# Week 3：Provider 抽象层与 Quota 基础闭环

## 一、目标说明

在 Week 2 已完成的基础设施（配置系统、日志系统、统一响应结构、数据库表结构等）之上，本周重点完成：

* Provider 抽象层设计与实现
* Provider Registry 管理机制
* Quota（容量）基础链路打通
* 最小可用 API 闭环

本周的核心目标是：

> 系统能够注册 provider、获取 provider 实例、同步 quota、并通过 API 查询 quota。

---

## 二、本周范围（必须完成）

### 2.1 Provider 抽象层

定义统一接口：

```go
type Provider interface {
    Name() string
    GetQuota(ctx context.Context, account *entity.ProviderAccount) (domain.Quota, error)
    GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error)
    RefreshToken(ctx context.Context, account *entity.ProviderAccount) error
}
```

### 设计说明

* `Name()`：用于注册表识别与日志输出
* `context.Context`：支持超时控制、链路追踪、取消请求
* `account` 显式传入：避免 provider 依赖全局变量，增强可测试性

---

### 2.2 Provider Registry

实现 Provider 注册与查找机制：

```go
type Registry struct {
    providers map[string]Provider
}
```

#### 必须实现的方法：

```go
func (r *Registry) Register(name string, p Provider) error
func (r *Registry) Get(name string) (Provider, bool)
func (r *Registry) MustGet(name string) Provider
func (r *Registry) List() []string
```

#### 要求

* 禁止使用 if-else 进行 provider 分发
* provider 必须通过 Registry 注册
* Registry 作为全局唯一 provider 管理入口

---

### 2.3 第一个 Provider（MockProvider）

本周仅实现一个 Provider，建议使用 MockProvider：

#### 返回固定数据：

```json
{
  "provider": "mock",
  "total": 1000,
  "used": 200,
  "available": 800
}
```

#### 实现要求：

```go
func (p *Provider) GetQuota(...) {
    return 固定 quota
}
```

* `GetDirectLink()`：可返回未实现错误
* `RefreshToken()`：可为空实现

#### 目的

* 验证 Provider 抽象设计
* 支撑 API 联调
* 不依赖真实网盘

---

## 三、Quota 模块

### 3.1 领域模型

```go
type Quota struct {
    Provider  string    `json:"provider"`
    Total     int64     `json:"total"`
    Used      int64     `json:"used"`
    Available int64     `json:"available"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

---

### 3.2 数据库存储设计

#### 当前值：使用 `ProviderAccount`

字段：

* `TotalQuota`
* `UsedQuota`
* `AvailableQuota`
* `LastQuotaSyncAt`
* `LastError`

#### 历史记录：新增 `QuotaSnapshot`

```go
type QuotaSnapshot struct {
    ID                uint `gorm:"primaryKey"`
    ProviderAccountID uint `gorm:"index;not null"`
    ProviderType      string `gorm:"size:50;index"`
    TotalQuota        int64
    UsedQuota         int64
    AvailableQuota    int64
    SyncStatus        string `gorm:"size:20;not null;default:'success'"`
    ErrorMessage      string `gorm:"type:text"`
    CreatedAt         time.Time
}
```

#### 数据分层说明

| 表名              | 用途       |
| --------------- | -------- |
| ProviderAccount | 当前 quota |
| QuotaSnapshot   | 历史记录     |

---

## 四、API 设计

---

### 4.1 POST `/api/v1/quota/query`

#### 请求

```json
{
  "provider": "mock"
}
```

#### 行为

* 查询数据库当前 quota
* 不触发远端调用

#### 返回

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "provider": "mock",
    "total": 1000,
    "used": 200,
    "available": 800,
    "updated_at": "2026-04-12T10:00:00Z"
  }
}
```

---

### 4.2 POST `/api/v1/quota/sync`

#### 请求

```json
{
  "provider": "mock"
}
```

#### 行为

* 查找 provider
* 调用 `GetQuota`
* 更新 ProviderAccount
* 写入 QuotaSnapshot

#### 返回

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "provider": "mock",
    "total": 1000,
    "used": 200,
    "available": 800,
    "updated_at": "2026-04-12T10:05:00Z"
  }
}
```

---

### 4.3 GET `/api/v1/providers`（建议实现）

#### 返回

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "name": "mock",
        "enabled": true
      }
    ]
  }
}
```

---

## 五、推荐目录结构

```text
internal/
  provider/
    provider.go
    registry.go
    mock/provider.go
  domain/
    entity/
      provider_account.go
      quota_snapshot.go
    dto/
      quota.go
  repository/
    provider_account_repository.go
    quota_snapshot_repository.go
  service/
    quota_service.go
  handler/
    quota_handler.go
  router/
    router.go
```

---

## 六、核心流程

### 6.1 Quota Sync 流程

```
API 请求
  ↓
Handler
  ↓
QuotaService
  ↓
Registry 获取 Provider
  ↓
Provider.GetQuota()
  ↓
更新 ProviderAccount
  ↓
写入 QuotaSnapshot
  ↓
返回结果
```

---

## 七、开发计划（建议）

| 天数    | 内容                        |
| ----- | ------------------------- |
| Day 1 | Provider 接口 + Registry    |
| Day 2 | MockProvider + Repository |
| Day 3 | QuotaService              |
| Day 4 | API Handler               |
| Day 5 | 测试与文档                     |

---

## 八、验收标准

* [ ] Provider 接口定义完成
* [ ] Registry 可注册/获取 provider
* [ ] MockProvider 可返回 quota
* [ ] `/quota/query` 可用
* [ ] `/quota/sync` 可用
* [ ] quota 写入 ProviderAccount
* [ ] quota 写入 QuotaSnapshot
* [ ] API 符合统一返回结构
* [ ] 错误码统一处理

---

## 九、本周不做

* 多 provider 支持
* 真实网盘接入
* 复杂 token 刷新
* 下载链路解析
* 前端复杂页面

---

## 十、注意事项

* 禁止 if-else 分发 provider
* provider 必须无状态
* query 与 sync 逻辑必须分离
* 所有 API 使用统一响应结构
* 必须记录错误日志

---
