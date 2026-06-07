# Week 4 - Baidu Provider & Quota Policy MVP

## 本周目标

在已有 Provider 与 Quota 基础结构上，引入第一个真实 Provider（百度网盘），并实现统一的容量策略模型（Quota Policy），支持三种容量模式：

* real（真实容量）
* inherit（向上继承容量）
* virtual（虚拟容量）

本周的核心目标是：

**将 Quota 层从“数据获取”升级为“容量策略解析层”**

并完成第一个完整可运行的 Provider + Quota 闭环。

---

## 核心里程碑

实现以下完整流程：

```text
创建 ProviderAccount（百度网盘）
→ 创建 MountPoint（挂载点 + quota 模式）
→ 获取真实容量（Provider）
→ 通过 QuotaResolver 计算最终容量
→ 返回统一 quota 结果
→ 写入 QuotaSnapshot
```

---

## 模块任务

---

### 任务一：BaiduProvider（必须完成）

#### 目标

实现第一个真实 Provider，用于获取百度网盘真实容量。

#### 要求

实现接口：

```go
type Provider interface {
    GetRealQuota(ctx context.Context, account ProviderAccount) (QuotaInfo, error)
}
```

#### BaiduProvider 必须支持：

* 获取真实 total / used / available
* 返回标准化 QuotaInfo
* 基础错误处理（token失效 / 请求失败）

#### 本周不要求：

* token 自动刷新
* 下载直链解析
* cookie 管理

---

### 任务二：QuotaMode 定义

定义统一容量模式：

```go
type QuotaMode string

const (
    QuotaModeReal    QuotaMode = "real"
    QuotaModeInherit QuotaMode = "inherit"
    QuotaModeVirtual QuotaMode = "virtual"
)
```

---

### 任务三：MountPoint 模型

#### 目标

引入挂载点级别配置，使 quota 模式与 provider 分离。

#### 数据结构

```go
type MountPoint struct {
    ID                uint
    Name              string
    ProviderAccountID uint
    ProviderType      string

    MountPath         string
    ProviderRootPath  string

    QuotaMode         string
    InheritFromID     *uint

    VirtualTotal      int64
    VirtualUsed       int64

    ReadOnly          bool
    Enabled           bool
}
```

---

### 任务四：QuotaResolver

#### 目标

实现统一容量解析逻辑：

```go
ResolveQuota(mountPoint) -> QuotaInfo
```

---

### 三种模式实现

#### 1. real

```text
total = provider.total
used = provider.used
available = total - used
```

---

#### 2. inherit

```text
total = parent.total
used = parent.used
available = parent.available
```

#### 约束：

* inherit 只能指向 real 模式挂载点
* 禁止循环继承

---

#### 3. virtual

```text
total = virtual_total
used = virtual_used
available = total - used
```

#### 约束：

* virtual_total <= allowed_max
* virtual_used <= virtual_total
* virtual_total = 0 → 只读模式

---

### allowed_max 定义

```text
allowed_max = 对应 root provider 的真实 total
```

---

### 任务五：QuotaSnapshot 扩展

#### 目标

记录每次 quota 计算结果

```go
type QuotaSnapshot struct {
    ID                uint
    MountPointID      uint
    ProviderAccountID uint

    Mode              string
    Total             int64
    Used              int64
    Available         int64

    SyncStatus        string
    ErrorMessage      string
    SyncedAt          time.Time
}
```

---

### 任务六：API 实现

#### 必须实现接口：

```text
POST /api/v1/mount
GET  /api/v1/mount/:id/quota
POST /api/v1/mount/:id/quota/sync
```

---

### 任务七：校验逻辑

#### 必须校验：

* inherit 不允许循环
* inherit 必须指向存在 mount
* virtual 不允许超过 allowed_max
* virtual_used <= virtual_total
* real 必须绑定 provider account

---

### 任务八：日志与调试

记录：

```text
provider 返回值
quota 解析模式
inherit 链
virtual 校验结果
最终 quota 输出
错误信息
```



##  验收标准

### 功能验收

* 能创建 BaiduProvider 账户
* 能创建 MountPoint 并选择 quota 模式
* 能返回三种模式的正确 quota

---

### 演示验收

#### 场景1：real

创建根目录挂载点 → 返回真实容量

#### 场景2：inherit

创建子目录 → 返回与 root 相同容量

#### 场景3：virtual

创建虚拟容量（如 100GB） → 正确返回
尝试超限 → 被拒绝

---

## 风险与注意事项

* 百度网盘 API 稳定性与 token 有效性
* inherit 可能出现错误引用或循环
* virtual 若未限制 allowed_max 将破坏系统一致性
* quota 失败必须可观测（不可静默失败）

---


