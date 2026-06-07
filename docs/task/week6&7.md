# Week6 开发任务说明：OpenList 存储驱动 & 文件列表能力集成

## 一、任务背景

在上周已实现本地 Disk Provider 支持，本周目标是进一步支持 **从 OpenList 获取各类存储驱动信息 (driver)**，并基于驱动类型处理 provider 激活逻辑，同时实现文件列表与文件详情的统一接口包裹层，为前端和内部业务提供稳定、统一的文件访问能力。

OpenList 是一个用于存储抽象与操作的开源服务。我们通过其 API 获取底层 provider 驱动信息与文件数据，并在当前项目进行二次加工封装。

## 二、本周核心目标

### 核心功能

1. **从 OpenList API 获取全部 driver 信息**
2. **根据 driver 类型决定是否需要二次激活 provider**
3. **通过 OpenList API 获取文件列表 & 文件信息**
4. **将 OpenList 原始 API 包装成我们内部统一的 API**
5. **完善错误处理、状态/日志记录等**

## 三、功能设计

### 3.1 Driver 信息获取

从 OpenList API 获取 driver 列表，例如：

```go
// 示例伪码
func FetchOpenListDrivers(ctx context.Context) ([]OpenListDriver, error) {
    resp, err := httpClient.Get("/openlist/api/v1/drivers")
    ...
}
```

需要返回的数据字段至少包括：

* driver id
* driver 类型 (e.g., local, s3, cloud, remote)
* 状态
* 是否已激活

### 3.2 Provider 二次激活策略

根据 driver 类型判断是否需要激活 provider：

| Driver 类型 | 是否需要二次激活      |
| --------- | ------------- |
| 本地 disk   | ❌ 不需要         |
| 第三方 云盘    | ✅ 需要 token/刷新 |
| 其他挂载      | 判断 mount 配置   |

激活逻辑示例伪码：

```go
switch driver.Type {
case "cloud":
    provider, err := ActivateProvider(driver)
    ...
default:
    // no-op
}
```

### 3.3 文件列表与信息获取

封装以下能力：

* 获取目录下文件列表
* 获取单个文件的元信息
* 统一校验和错误处理

示例：

```go
func ListOpenListFiles(ctx context.Context, driverID string, path string) ([]FileInfo, error) {
    // 调用 OpenList API /files?driver={driverID}&path={path}
}

func GetOpenListFileInfo(ctx context.Context, driverID, fileID string) (FileInfo, error) {
    // 调用 OpenList API /file/{fileID}
}
```

### 3.4 内部 API 统一包装

对外提供统一接口，例如：

```
GET /api/v1/storage/{driverID}/files?path=/xxx
GET /api/v1/storage/{driverID}/file/{fileID}
```

返回 JSON 示例：

```json
{
  "driver": "my-driver-id",
  "path": "/",
  "files": [
    {
      "name": "file1.txt",
      "id": "id123",
      "size": 12345,
      "type": "file"
    }
  ]
}
```

## 四、接口定义

### 4.1 获取所有 Driver

```
GET /api/v1/storage/drivers
```

Response:

```json
[
  {
    "id": "driver1",
    "type": "cloud",
    "active": true
  }
]
```

### 4.2 列出文件

```
GET /api/v1/storage/{driverID}/files?path=/
```

Response:

```json
{
  "path": "/",
  "files": [...]
}
```

### 4.3 文件信息

```
GET /api/v1/storage/{driverID}/file/{fileID}
```

Response:

```json
{
  "id": "fileID",
  "name": "file.png",
  "size": 2048
}
```

## 五、错误处理 & 日志设计

系统需保证：

* 对 OpenList API 调用失败进行重试策略
* 若 provider 激活失败，需明确错误返回给前端
* 必要的日志记录：调用参数、返回状态码与错误信息


## 六、本周验收标准

符合以下条件即通过：

* 成功调用 OpenList API 并返回 driver 列表
* 按驱动类型正确处理 provider 激活
* 文件列表与信息接口可正常使用
* 对外统一 API 返回正确 JSON
* 无明显崩溃与未处理异常
