# Week10 开发任务说明：夸克网盘挂载解析、容量返回、默认配置恢复与 EXE 打包启动

## 一、任务背景

Week6&7 已完成 OpenList 存储驱动、文件列表与文件详情能力封装。

Week8&9 已完成直链解析与 aria2 下载提交闭环。

本周目标是在现有 OpenBridge 架构上，新增：

**夸克网盘 Quark Provider 支持**

并完成以下能力：

1. 解析 OpenList 中的夸克网盘挂载
2. 返回夸克网盘容量信息
3. 将夸克网盘接入统一 Provider / Mount / Quota 体系
4. 前端展示夸克挂载与容量
5. 支持一键恢复默认配置文件
6. 支持项目打包为 Windows `.exe` 可执行文件
7. 支持通过 `./openbridge.exe server -port` 指令启动 Server，默认端口为 `8080`
8. 实现普通用户与管理员用户权限分离
9. 只有管理员可以删除挂载存储、修改配置和恢复默认配置
10. 普通用户只能查看挂载、容量和文件信息，不能删除和配置

---

## 二、本周核心目标

### 核心功能

1. 新增 `net_disk = "quark"`
2. 支持识别 OpenList driver / mount 中的夸克网盘
3. 支持夸克网盘容量查询
4. 支持夸克网盘挂载点解析
5. 支持 quota real / inherit / virtual 模式
6. 提供一键恢复默认配置接口
7. 前端增加恢复默认配置按钮
8. 支持 Windows EXE 打包输出
9. 支持命令行启动 Server

---

## 三、关键机制说明

### 3.1 夸克网盘数据来源

夸克网盘本身由 OpenList 负责挂载。

OpenBridge 不直接管理夸克网盘文件流，只负责：

```text
OpenList 获取挂载信息
OpenBridge 解析 driver / mount
OpenBridge 查询容量
OpenBridge 返回统一 quota
前端展示容量与挂载状态
```

---

### 3.2 正确的数据流

```text
前端
 ↓
OpenBridge API
 ↓
StorageUseCase / QuotaUseCase
 ↓
OpenList Client
 ↓
Quark Mount / Driver
 ↓
返回统一容量结构
```

---

### 3.3 本周不负责

* 不实现夸克网页登录
* 不保存用户账号密码
* 不直接下载夸克文件流
* 不绕过 OpenList 访问夸克私有接口

---

## 四、后端任务

## 任务一：新增 Quark Provider

新增文件：

```text
backend/internal/domain/providers/quark_provider.go
```

实现统一 Provider 接口：

```go
type QuarkProvider struct {
    openListClient *openlist.Client
}

func NewQuarkProvider(client *openlist.Client) *QuarkProvider {
    return &QuarkProvider{
        openListClient: client,
    }
}

func (p *QuarkProvider) Name() string {
    return "quark"
}

func (p *QuarkProvider) GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error) {
    // 从 OpenList driver / mount 信息中解析夸克容量
}

func (p *QuarkProvider) GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error) {
    // 可复用 Week8&9 的直链解析逻辑
}

func (p *QuarkProvider) RefreshToken(ctx context.Context, account *entity.ProviderAccount) error {
    // 当前阶段由 OpenList 管理 token，OpenBridge 暂不刷新
    return nil
}
```

---

## 任务二：夸克挂载解析

新增结构：

```go
type QuarkMountInfo struct {
    MountID      string `json:"mount_id"`
    MountPath    string `json:"mount_path"`
    DriverType   string `json:"driver_type"`
    Provider     string `json:"provider"`
    Status       string `json:"status"`
    Total        int64  `json:"total"`
    Used         int64  `json:"used"`
    Available    int64  `json:"available"`
    IsAvailable  bool   `json:"is_available"`
}
```

新增方法：

```go
func (s *StorageUseCase) ResolveQuarkMounts(ctx context.Context) ([]QuarkMountInfo, error)
```

逻辑：

```text
1. 调用 OpenList driver / storage API
2. 遍历所有挂载点
3. 判断 driver 类型是否为 quark
4. 读取挂载路径、状态、容量字段
5. 转换为 OpenBridge 内部统一结构
6. 返回给前端
```

---

## 任务三：容量返回

统一返回结构：

```go
type Quota struct {
    Provider  string    `json:"provider"`
    Total     int64     `json:"total"`
    Used      int64     `json:"used"`
    Available int64     `json:"available"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

夸克容量计算规则：

```text
total = quark.total
used = quark.used
available = total - used
```

异常处理：

```text
如果 OpenList 未返回 total：
    total = 0

如果 OpenList 未返回 used：
    used = 0

如果 available < 0：
    available = 0

如果挂载不可用：
    status = error
    last_error 写入 ProviderAccount
```

---

## 任务四：Provider 注册

在 Provider Registry 中注册：

```go
registry.Register("quark", providers.NewQuarkProvider(openListClient))
```

禁止通过大量 if-else 分发 provider。

---

## 任务五：MountPoint 支持 Quark

创建夸克挂载点示例：

```json
{
  "name": "quark-root",
  "provider_type": "quark",
  "net_disk": "quark",
  "mount_path": "/quark",
  "provider_root_path": "/",
  "quota_mode": "real"
}
```

要求：

* `quota_mode = real` 时返回夸克真实容量
* `quota_mode = inherit` 时继承父级容量
* `quota_mode = virtual` 时返回虚拟容量
* 挂载不可用时必须返回明确错误

---

## 任务六：一键恢复默认配置文件

### 6.1 默认配置文件

新增：

```text
backend/config/default.env
```

示例：

```env
APP_NAME=OpenBridge
APP_ENV=dev
APP_PORT=8080

DB_PATH=./data/openbridge.db

ARIA2_RPC_URL=http://127.0.0.1:6800/jsonrpc
ARIA2_RPC_SECRET=
ARIA2_DOWNLOAD_DIR=./downloads

OPENLIST_BASE_URL=http://127.0.0.1:5244
OPENLIST_TOKEN=

LOG_LEVEL=debug
LOG_FORMAT=json

DEFAULT_PROVIDER=quark
```

---

### 6.2 配置恢复逻辑

新增：

```go
func (s *ConfigUseCase) RestoreDefaultConfig(ctx context.Context) error
```

流程：

```text
1. 读取 backend/config/default.env
2. 备份当前 .env 为 .env.backup
3. 用 default.env 覆盖 .env
4. 重新加载配置
5. 返回恢复结果
```

---

### 6.3 API 设计

#### 恢复默认配置

```http
POST /api/v1/config/restore-default
```

返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "restored": true,
    "backup_file": ".env.backup"
  }
}
```

#### 查询当前配置

```http
GET /api/v1/config
```

注意：

* 不返回 token 明文
* secret 字段必须脱敏

示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "app_env": "dev",
    "openlist_base_url": "http://127.0.0.1:5244",
    "openlist_token": "******",
    "aria2_rpc_url": "http://127.0.0.1:6800/jsonrpc"
  }
}
```

---

## 任务七：普通用户与管理员权限分离

系统需要区分两类用户：

```text
admin  管理员用户
user   普通用户
```

权限要求：

| 用户角色  | 查看挂载 | 查看容量 | 查看文件 | 删除挂载存储 | 修改配置 | 恢复默认配置 |
| ----- | ---- | ---- | ---- | ------ | ---- | ------ |
| admin | 允许   | 允许   | 允许   | 允许     | 允许   | 允许     |
| user  | 允许   | 允许   | 允许   | 禁止     | 禁止   | 禁止     |

---

### 7.1 用户角色字段

用户实体需要增加角色字段：

```go
type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Role     string `json:"role"`
}
```

角色枚举：

```go
const (
    RoleAdmin = "admin"
    RoleUser  = "user"
)
```

默认规则：

```text
首次初始化系统时创建 admin 用户
后续新建用户默认角色为 user
禁止普通用户修改自己的 role
```

---

### 7.2 权限中间件

新增管理员权限校验中间件：

```go
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := CurrentUser(c)
        if user == nil || user.Role != "admin" {
            c.JSON(403, gin.H{
                "code": 403,
                "message": "admin permission required",
                "data": nil,
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

普通登录用户只需要通过认证中间件：

```go
func RequireAuth() gin.HandlerFunc
```

---

### 7.3 接口权限划分

只允许管理员访问的接口：

```http
DELETE /api/v1/storage/mounts/:id
POST   /api/v1/config/restore-default
PUT    /api/v1/config
POST   /api/v1/storage/mounts
PUT    /api/v1/storage/mounts/:id
DELETE /api/v1/provider/accounts/:id
```

普通用户允许访问的接口：

```http
GET  /api/v1/storage/quark/mounts
POST /api/v1/quota/query
GET  /api/v1/config
GET  /api/v1/storage/files
GET  /api/v1/storage/detail
```

注意：

* 普通用户访问删除、修改、配置接口时必须返回 `403`
* 普通用户不应该看到敏感配置原文
* 管理员接口必须同时经过登录认证和管理员权限校验
* 前端隐藏按钮只是体验优化，后端接口必须强制校验权限

---

### 7.4 删除挂载存储权限

删除挂载存储接口必须强制管理员权限：

```http
DELETE /api/v1/storage/mounts/:id
```

处理逻辑：

```text
1. 校验用户是否已登录
2. 校验用户 role 是否为 admin
3. 如果不是 admin，返回 403
4. 如果是 admin，继续删除挂载存储
5. 记录操作日志
```

返回示例：

```json
{
  "code": 403,
  "message": "admin permission required",
  "data": null
}
```

---

### 7.5 配置权限

以下配置操作仅管理员可用：

```http
PUT  /api/v1/config
POST /api/v1/config/restore-default
```

普通用户只能查看脱敏后的配置：

```http
GET /api/v1/config
```

脱敏要求：

```text
OPENLIST_TOKEN      => ******
ARIA2_RPC_SECRET    => ******
cookie              => ******
authorization       => ******
```

---

### 7.6 前端权限控制

前端需要根据当前登录用户角色控制页面能力。

管理员可见：

* 删除挂载按钮
* 新增挂载按钮
* 编辑挂载按钮
* 恢复默认配置按钮
* 配置保存按钮

普通用户不可见：

* 删除挂载按钮
* 新增挂载按钮
* 编辑挂载按钮
* 恢复默认配置按钮
* 配置保存按钮

普通用户页面只展示：

```text
挂载列表
容量信息
文件列表
文件详情
```

---

### 7.7 操作日志

管理员执行敏感操作时必须记录日志：

```text
admin delete mount
admin update config
admin restore default config
admin create mount
admin update mount
```

日志字段：

```text
user_id
username
role
action
resource_id
created_at
```

日志中禁止出现 token、secret、cookie、authorization。

---

### 7.8 验收标准

* [ ] 用户实体包含 `role` 字段
* [ ] 系统支持 `admin` 和 `user` 两种角色
* [ ] 管理员可以删除挂载存储
* [ ] 普通用户不能删除挂载存储
* [ ] 普通用户访问删除接口返回 `403`
* [ ] 管理员可以修改配置和恢复默认配置
* [ ] 普通用户不能修改配置和恢复默认配置
* [ ] 普通用户只能查看挂载、容量和文件信息
* [ ] 前端根据角色隐藏管理按钮
* [ ] 后端接口强制校验管理员权限
* [ ] 敏感配置对普通用户和管理员返回时均需脱敏

---

## 五、命令行启动与 EXE 打包任务

## 任务七：支持打包为 Windows `.exe`

项目需要支持将后端服务打包为 Windows 可执行文件：

```text
openbridge.exe
```

推荐输出目录：

```text
release/
  openbridge.exe
  config/
    default.env
  data/
  downloads/
```

---

### 7.1 打包命令

在项目根目录执行：

```bash
cd backend
GOOS=windows GOARCH=amd64 go build -o ../release/openbridge.exe ./cmd/openbridge
```

如果当前系统为 Windows，可直接执行：

```bash
cd backend
go build -o ../release/openbridge.exe ./cmd/openbridge
```

---

### 7.2 打包脚本

新增 Windows 打包脚本：

```text
scripts/build-windows.bat
```

内容示例：

```bat
@echo off
setlocal

set APP_NAME=openbridge.exe
set OUTPUT_DIR=release

if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%
if not exist %OUTPUT_DIR%\config mkdir %OUTPUT_DIR%\config
if not exist %OUTPUT_DIR%\data mkdir %OUTPUT_DIR%\data
if not exist %OUTPUT_DIR%\downloads mkdir %OUTPUT_DIR%\downloads

cd backend

go build -o ..\%OUTPUT_DIR%\%APP_NAME% .\cmd\openbridge

cd ..

copy backend\config\default.env %OUTPUT_DIR%\config\default.env

echo Build success: %OUTPUT_DIR%\%APP_NAME%
endlocal
```

新增 Linux / macOS 打包脚本：

```text
scripts/build-windows.sh
```

内容示例：

```bash
#!/usr/bin/env bash
set -e

APP_NAME="openbridge.exe"
OUTPUT_DIR="release"

mkdir -p "$OUTPUT_DIR/config"
mkdir -p "$OUTPUT_DIR/data"
mkdir -p "$OUTPUT_DIR/downloads"

cd backend
GOOS=windows GOARCH=amd64 go build -o "../$OUTPUT_DIR/$APP_NAME" ./cmd/openbridge
cd ..

cp backend/config/default.env "$OUTPUT_DIR/config/default.env"

echo "Build success: $OUTPUT_DIR/$APP_NAME"
```

---

## 任务八：支持命令行启动 Server

需要支持以下命令启动服务：

```bash
./openbridge.exe server
```

默认监听端口：

```text
8080
```

也需要支持通过端口参数启动：

```bash
./openbridge.exe server -port 9090
```

或：

```bash
./openbridge.exe server --port 9090
```

---

### 8.1 命令行设计

推荐命令结构：

```text
openbridge.exe
  server
    -port / --port
```

示例：

```bash
./openbridge.exe server
./openbridge.exe server -port 8080
./openbridge.exe server --port 9090
```

---

### 8.2 默认端口优先级

端口读取优先级：

```text
命令行 -port / --port
 ↓
环境变量 APP_PORT
 ↓
配置文件 APP_PORT
 ↓
默认值 8080
```

也就是说：

```bash
./openbridge.exe server -port 9090
```

优先使用 `9090`。

如果未传端口：

```bash
./openbridge.exe server
```

默认使用 `8080`。

---

### 8.3 命令行入口示例

新增或调整入口：

```text
backend/cmd/openbridge/main.go
```

示例代码：

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "openbridge/internal/bootstrap"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    switch os.Args[1] {
    case "server":
        runServer(os.Args[2:])
    case "help", "-h", "--help":
        printUsage()
    default:
        fmt.Printf("unknown command: %s\n", os.Args[1])
        printUsage()
        os.Exit(1)
    }
}

func runServer(args []string) {
    fs := flag.NewFlagSet("server", flag.ExitOnError)
    port := fs.Int("port", 8080, "server listen port")

    _ = fs.Parse(args)

    cfg, err := bootstrap.LoadConfig()
    if err != nil {
        fmt.Printf("load config failed: %v\n", err)
        os.Exit(1)
    }

    if *port != 8080 {
        cfg.App.Port = *port
    }

    if cfg.App.Port == 0 {
        cfg.App.Port = 8080
    }

    if err := bootstrap.RunServer(cfg); err != nil {
        fmt.Printf("server start failed: %v\n", err)
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println(`OpenBridge

Usage:
  openbridge.exe server
  openbridge.exe server -port 8080
  openbridge.exe server --port 9090

Commands:
  server    Start OpenBridge server
`)
}
```

注意：

* `server` 是必需子命令
* `-port` 默认为 `8080`
* `--port` 需要兼容
* 端口非法时应返回明确错误
* 启动失败时应输出错误并退出

---

### 8.4 Server 启动日志

启动成功后输出：

```text
OpenBridge server starting...
Listen address: 0.0.0.0:8080
Config file: .env
```

如果使用自定义端口：

```text
OpenBridge server starting...
Listen address: 0.0.0.0:9090
Config file: .env
```

---

### 8.5 运行目录要求

`openbridge.exe` 启动时应支持以下目录结构：

```text
release/
  openbridge.exe
  .env
  config/
    default.env
  data/
    openbridge.db
  downloads/
```

如果 `.env` 不存在：

```text
1. 检查 config/default.env 是否存在
2. 如果存在，则复制为 .env
3. 如果不存在，则使用内置默认配置
4. 输出提示日志
```

---

## 六、接口定义

### 6.1 获取夸克挂载

```http
GET /api/v1/storage/quark/mounts
```

返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "mount_id": "quark-001",
        "mount_path": "/quark",
        "driver_type": "quark",
        "provider": "quark",
        "status": "active",
        "total": 1099511627776,
        "used": 214748364800,
        "available": 884763262976,
        "is_available": true
      }
    ]
  }
}
```

---

### 6.2 同步夸克容量

```http
POST /api/v1/quota/sync
```

请求：

```json
{
  "provider": "quark"
}
```

返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "provider": "quark",
    "total": 1099511627776,
    "used": 214748364800,
    "available": 884763262976,
    "updated_at": "2026-06-01T10:00:00Z"
  }
}
```

---

### 6.3 查询夸克容量

```http
POST /api/v1/quota/query
```

请求：

```json
{
  "provider": "quark"
}
```

---

### 6.4 恢复默认配置

```http
POST /api/v1/config/restore-default
```

---

## 七、前端任务

### 7.1 夸克挂载列表

新增页面或模块：

```text
QuarkMountPanel.vue
```

展示字段：

* 挂载路径
* 状态
* 总容量
* 已用容量
* 可用容量
* 是否可用

---

### 7.2 容量展示

显示：

```text
总容量：xxx GB
已用：xxx GB
可用：xxx GB
使用率：xx%
```

---

### 7.3 一键恢复默认配置按钮

新增按钮：

```text
恢复默认配置
```

点击流程：

```text
点击按钮
 ↓
弹出确认框
 ↓
调用 /api/v1/config/restore-default
 ↓
提示恢复成功
 ↓
刷新当前配置
```

确认文案：

```text
恢复默认配置会覆盖当前 .env，并自动生成备份文件。是否继续？
```

---

## 八、错误处理

必须处理以下情况：

| 场景                           | 处理                                 |
| ---------------------------- | ---------------------------------- |
| OpenList 未启动                 | 返回 OpenList 连接失败                   |
| OpenList Token 无效            | 返回未授权                              |
| 未找到 quark 挂载                 | 返回空列表                              |
| quark 挂载异常                   | status = error                     |
| 容量字段缺失                       | 使用默认值 0                            |
| 配置恢复失败                       | 返回明确错误                             |
| `.env` 备份失败                  | 禁止覆盖原配置                            |
| `openbridge.exe server` 启动失败 | 输出错误并退出                            |
| 端口被占用                        | 输出端口占用错误                           |
| 端口参数非法                       | 输出非法端口错误                           |
| 普通用户删除挂载                     | 返回 `403 admin permission required` |
| 普通用户修改配置                     | 返回 `403 admin permission required` |
| 普通用户恢复默认配置                   | 返回 `403 admin permission required` |

---

## 九、日志要求

必须记录：

```text
quark mount resolve start
quark mount count
quark quota sync start
quark quota result
restore default config start
backup env success
restore default config success
restore default config failed
openbridge build start
openbridge build success
openbridge server start
openbridge server listen port
openbridge server stop
admin delete mount
admin update config
admin restore default config
permission denied
```

日志中禁止输出：

```text
OPENLIST_TOKEN
ARIA2_RPC_SECRET
cookie
authorization
```

---

## 十、验收标准

满足以下条件即通过：

* [ ] 能识别 OpenList 中的夸克挂载
* [ ] 能返回夸克挂载列表
* [ ] 能同步夸克容量
* [ ] 能查询夸克容量
* [ ] quota real 模式可用
* [ ] quota inherit / virtual 不受影响
* [ ] 前端能展示夸克容量
* [ ] 一键恢复默认配置可用
* [ ] 恢复前会自动备份 `.env`
* [ ] token / secret 不会明文返回前端
* [ ] API 仍符合统一响应结构
* [ ] 能成功打包生成 `openbridge.exe`
* [ ] `./openbridge.exe server` 能以默认端口 `8080` 启动
* [ ] `./openbridge.exe server -port 9090` 能以指定端口启动
* [ ] `./openbridge.exe server --port 9090` 能以指定端口启动
* [ ] `.env` 不存在时能自动从 `config/default.env` 初始化
* [ ] 管理员可以删除挂载存储
* [ ] 普通用户不能删除挂载存储
* [ ] 普通用户不能修改配置
* [ ] 普通用户不能恢复默认配置
* [ ] 普通用户只能查看挂载、容量和文件信息

---

## 十一、测试用例

### 11.1 查询夸克挂载

```bash
curl http://localhost:8080/api/v1/storage/quark/mounts
```

---

### 11.2 同步夸克容量

```bash
curl -X POST http://localhost:8080/api/v1/quota/sync \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "quark"
  }'
```

---

### 11.3 查询夸克容量

```bash
curl -X POST http://localhost:8080/api/v1/quota/query \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "quark"
  }'
```

---

### 11.4 恢复默认配置

```bash
curl -X POST http://localhost:8080/api/v1/config/restore-default
```

---

### 11.5 打包 Windows EXE

```bash
bash scripts/build-windows.sh
```

或在 Windows 中执行：

```bat
scripts\build-windows.bat
```

预期输出：

```text
release/openbridge.exe
```

---

### 11.6 默认端口启动 Server

```bash
cd release
./openbridge.exe server
```

预期：

```text
Listen address: 0.0.0.0:8080
```

---

### 11.7 指定端口启动 Server

```bash
cd release
./openbridge.exe server -port 9090
```

预期：

```text
Listen address: 0.0.0.0:9090
```

---

### 11.8 双横线端口参数

```bash
cd release
./openbridge.exe server --port 9090
```

预期：

```text
Listen address: 0.0.0.0:9090
```

---

### 11.9 普通用户删除挂载失败

```bash
curl -X DELETE http://localhost:8080/api/v1/storage/mounts/quark-root \
  -H "Authorization: Bearer USER_TOKEN"
```

预期：

```json
{
  "code": 403,
  "message": "admin permission required",
  "data": null
}
```

---

### 11.10 管理员删除挂载成功

```bash
curl -X DELETE http://localhost:8080/api/v1/storage/mounts/quark-root \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

预期：

```json
{
  "code": 0,
  "message": "ok"
}
```

---

### 11.11 普通用户恢复默认配置失败

```bash
curl -X POST http://localhost:8080/api/v1/config/restore-default \
  -H "Authorization: Bearer USER_TOKEN"
```

预期：

```json
{
  "code": 403,
  "message": "admin permission required",
  "data": null
}
```

---

## 十二、本周输出

```text
backend/
  cmd/
    openbridge/
      main.go
  config/
    default.env
  internal/
    domain/
      providers/
        quark_provider.go
    usecase/
      storage_usecase.go
      quota_usecase.go
      config_usecase.go
    handler/
      quark_handler.go
      config_handler.go
    middleware/
      auth_middleware.go
      admin_middleware.go
    entity/
      user.go

frontend/
  src/
    views/
      QuarkMountPanel.vue
    components/
      RestoreDefaultConfigButton.vue

scripts/
  build-windows.bat
  build-windows.sh

release/
  openbridge.exe
  config/
    default.env
  data/
  downloads/
```

---

## 十三、注意事项

* Quark Provider 不应直接依赖全局配置
* OpenList Token 必须统一从 Config 读取
* 配置恢复必须先备份再覆盖
* 恢复配置后需要重新加载 Config
* 容量单位后端统一使用 byte
* 前端负责格式化 GB / TB
* 所有 API 返回必须符合统一响应结构
* `openbridge.exe server` 必须作为标准启动方式
* 默认端口必须为 `8080`
* 命令行端口参数优先级必须高于 `.env`
* 打包产物必须包含 `config/default.env`
* 日志中不得输出 token、secret、cookie 或 authorization
* 删除挂载、修改配置、恢复默认配置必须只有管理员可操作
* 普通用户只能查看，不能删除挂载，不能修改配置
