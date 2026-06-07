# Week8 开发任务说明：直链解析与 aria2 下载提交闭环

## 一、任务背景

Week6&7 已完成 OpenList driver、文件列表、文件信息接口封装，并且当前 `FileDetail` 已包含 `raw_url` 字段，可作为直链解析的基础。

Week8 的目标是打通：

```text
平板/前端选择文件
→ OpenBridge 请求 OpenList 获取文件信息
→ 解析 raw_url / direct_link
→ OpenBridge 调 aria2 JSON-RPC addUri
→ aria2 在挂载电脑本机下载
```

注意：OpenBridge 后端不负责转发文件流，只负责任务编排。

---

## 二、本周核心目标

1. 获取 OpenList 文件直链
2. 新增 aria2 RPC Client
3. 将直链提交给 aria2
4. 记录下载任务与 aria2 GID 映射
5. 提供任务查询接口
6. 保证下载流不经过 OpenBridge

---

## 三、关键机制说明（重点）

### 正确的数据流

```text
OpenBridge → 获取 direct_link
OpenBridge → aria2.addUri(direct_link)
aria2 → 直链源站下载
```

 文件数据不会经过 OpenBridge

---

### 可能走代理的情况

如果直链是：

```text
http://你的电脑:5244/d/xxx
```

说明：

 仍然经过 OpenList（但不是 OpenBridge）

---

### ✔ 必须做的判断

```json
"is_openlist_proxy": true / false
```

---

## 四、后端任务

### 直链解析

新增：

```go
type DirectLinkResult struct {
    Path            string
    Name            string
    Size            int64
    Provider        string
    DirectLink      string
    IsOpenListProxy bool
}
```

方法：

```go
func (s *StorageUseCase) ResolveDirectLink(ctx context.Context, path string) (*DirectLinkResult, error)
```

逻辑：

* 调 `/api/fs/get`
* 取 `raw_url`
* 判断 host 是否为 OpenList

---

### aria2 Client

目录：

```text
internal/aria2/client.go
```

核心方法：

```go
AddURI(uri string) (gid string, error)
TellStatus(gid string)
Remove(gid string)
```

---

### Download UseCase

```go
CreateTask(path, dir)
GetTask(taskID)
```

流程：

```text
Resolve → AddURI → 保存任务 → 返回 gid
```

---

### 数据库表

```go
type DownloadTask struct {
    TaskID     string
    SourcePath string
    DirectLink string
    FileName   string
    FileSize   int64
    Aria2GID   string
    Status     string
}
```

---

### API

#### 解析直链

```http
POST /download/resolve
```

---

#### 创建任务

```http
POST /download/tasks
```

---

#### 查询任务

```http
GET /download/tasks/:id
```

---

## 五、前端任务

### 文件列表增加按钮

```text
发送到 aria2
```

---

### 弹窗显示

* 文件名
* Provider
* DirectLink
* 是否代理

---

### 下载任务列表

显示：

* 状态
* 进度
* 速度

---

## 六、配置

```env
ARIA2_RPC_URL=http://127.0.0.1:6800/jsonrpc
ARIA2_RPC_SECRET=
ARIA2_DOWNLOAD_DIR=D:/Downloads
OPENLIST_BASE_URL=http://127.0.0.1:5244
OPENLIST_TOKEN=
```

---

## 七、验收标准

* 能获取直链
* 能识别代理
* 能提交 aria2
* 能返回 gid
* 能查询状态
* 下载在本机 aria2 执行
* OpenBridge 不参与数据传输



