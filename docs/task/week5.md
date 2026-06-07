# Week5 开发任务说明：本地 Disk Provider 支持实现

## 一、任务背景

在 Week4 中，系统已完成容量适配核心功能，支持：

* mock provider（模拟容量）
* 百度网盘 provider（真实容量）
* MountPoint 三种模式（real / inherit / virtual）

当前系统已经具备基础可用性

因此，本周引入：

 **本地 Disk Provider

---

## 二、本周目标

###  核心目标

实现一个新的 provider 类型：

```text
net_disk = "local"
```

并完成以下能力：

1. 获取本地磁盘容量（total / used / available）
2. 接入现有 Provider 架构（统一接口）
3. 支持 quota 查询与 sync
4. 支持 MountPoint real 模式
5. 可用于前端展示与系统演示

---

## 三、功能设计

### 3.1 Local Provider 定义

新增文件：

```
backend/internal/domain/providers/local_provider.go
```

实现接口：

```go
type Provider interface {
    GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error)
    GetDirectLink(ctx context.Context, fileID string) (string, error)
    RefreshToken(ctx context.Context, account *entity.ProviderAccount) error
}
```

---

### 3.2 容量获取方式

本地磁盘容量通过系统调用获取：

#### Linux / macOS

使用：

```go
syscall.Statfs
```

示例逻辑：

```go
var stat syscall.Statfs_t
syscall.Statfs("/", &stat)

total := stat.Blocks * uint64(stat.Bsize)
available := stat.Bavail * uint64(stat.Bsize)
used := total - available
```

---

### 3.3 返回结构

```json
{
  "provider": "local",
  "total": xxx,
  "used": xxx,
  "available": xxx
}
```

---

### 3.4 Provider 注册

在：

```
provider_usecase.go
mount_usecase.go
quota_usecase.go
```

中增加：

```go
case "local":
    return providers.NewLocalProvider()
```

---

## 四、接口测试设计

### 4.1 注册 Local Provider

```bash
curl -X POST http://localhost:8080/api/v1/provider \
-H "Content-Type: application/json" \
-d '{
  "name": "local-provider",
  "net_disk": "local"
}'
```

---

### 4.2 同步本地容量

```bash
curl -X POST http://localhost:8080/api/v1/quota/sync \
-H "Content-Type: application/json" \
-d '{
  "name": "local-provider"
}'
```

---

### 4.3 创建 Mount（real）

```bash
curl -X POST http://localhost:8080/api/v1/mount \
-H "Content-Type: application/json" \
-d '{
  "name": "local-root",
  "provider_account_id": <id>,
  "mount_path": "/local",
  "provider_root_path": "/",
  "quota_mode": "real"
}'
```

---

### 4.4 查询 quota

```bash
curl http://localhost:8080/api/v1/mount/<id>/quota
```

---

## 五、预期效果

| 项目         | 预期           |
| ---------- | ------------ |
| quota/sync | 返回本机真实磁盘容量   |
| mount real | 使用本地磁盘作为容量来源 |
| 稳定性        | 不依赖网络，始终可用   |
| 演示效果       | 可展示真实磁盘占用变化  |

---

## 六、扩展方向

后续可扩展：

* 指定路径（不同磁盘分区）
* 多磁盘支持
* 挂载目录级容量统计
* 本地文件扫描统计 used（更精确）

---

## 七、分工建议

| 成员  | 任务                    |
| --- | --------------------- |
| 后端1 | 实现 local provider     |
| 后端2 | 接入 usecase 与 registry |
| 测试  | 编写接口测试用例              |
| 前端  | 增加 local provider 展示  |

---

## 九、本周验收标准

满足以下条件即通过：

* [ ] local provider 注册成功
* [ ] quota/sync 返回真实磁盘容量
* [ ] mount real 模式可正常使用
* [ ] 无崩溃 / panic
* [ ] 数据格式符合统一规范
