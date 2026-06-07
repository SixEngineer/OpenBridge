# OpenBridge API 文档（临时开发版）

> 版本：v0.1 (Draft)
> 更新时间：2026-03-28
> 适用对象：后端开发 / 前端开发 / 自动化脚本开发

---

# 一、设计原则

## 1.1 分层设计

API 按业务模块划分：

* 认证与管理员接口
* 系统状态接口
* OpenList 适配接口
* Provider 扩展接口
* 下载编排接口
* aria2 任务接口
* 容量（Quota）接口
* 配置接口
* 日志与调试接口

## 1.2 统一返回格式

### 成功

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### 失败

```json
{
  "code": 1001,
  "message": "error message",
  "data": null
}
```

## 1.3 通用约定

### 基础路径

```
/api/v1
```

### 鉴权方式

```
Authorization: Bearer <token>
```

### 分页参数

```json
{
  "page": 1,
  "page_size": 20
}
```

### 时间格式

```
ISO8601
```

---

# 二、认证 API

## 2.1 登录

**POST /api/v1/auth/login**

请求：

```json
{
  "username": "admin",
  "password": "123456"
}
```

返回：

```json
{
  "code": 0,
  "data": {
    "access_token": "xxx",
    "refresh_token": "xxx",
    "expires_in": 7200,
    "user": {
      "id": 1,
      "username": "admin",
      "role": "super_admin"
    }
  }
}
```

错误码：

* 1: wrong password
* 2: unknown user / error

---

## 2.2 登出

**POST /api/v1/auth/logout**

---

## 2.3 用户偏好设置

**PUT /api/v1/users/me/preferences**

```json
{
  "theme": "dark",
  "default_home": "downloads",
  "show_debug_info": true
}
```

---

# 三、系统状态 API

## 3.1 健康检查

**GET /api/v1/system/health**

## 3.2 仪表盘

**GET /api/v1/system/dashboard**

## 3.3 组件状态

**GET /api/v1/system/status**

---

# 四、OpenList 适配 API

## 4.1 测试连接

**POST /api/v1/openlist/test**

## 4.2 获取驱动列表

**GET /api/v1/openlist/drivers**

## 4.3 获取目录列表

**POST /api/v1/fs/list**

## 4.4 获取文件信息

**POST /api/v1/openlist/fs/get**

## 4.5 获取原始下载入口

**POST /api/v1/fs/raw-link**

---

# 五、Provider API（核心扩展层）

## 5.1 获取 Provider 列表

**GET /api/v1/providers**

## 5.2 Provider 详情

**GET /api/v1/providers/{provider_id}**

## 5.3 激活 Provider

**POST /api/v1/providers/{provider_id}/activate**

## 5.4 停用 Provider

**POST /api/v1/providers/{provider_id}/deactivate**

## 5.5 更新二次认证

**PUT /api/v1/providers/{provider_id}/secondary-auth**

## 5.6 测试二次认证

**POST /api/v1/providers/{provider_id}/secondary-auth/test**

## 5.7 能力查询

**GET /api/v1/providers/{provider_id}/capabilities**

---

# 六、下载编排 API

## 6.1 解析下载链路

**POST /api/v1/download/resolve**

## 6.2 提交下载任务

**POST /api/v1/download/tasks**

## 6.3 获取任务列表

**GET /api/v1/download/tasks**

## 6.4 任务详情

**GET /api/v1/download/tasks/{task_id}**

## 6.5 重试任务

**POST /api/v1/download/tasks/{task_id}/retry**

## 6.6 取消任务

**POST /api/v1/download/tasks/{task_id}/cancel**

## 6.7 删除任务

**DELETE /api/v1/download/tasks/{task_id}**

---

# 七、容量（Quota）API

## 7.1 查询路径容量

**POST /api/v1/quota/query**

## 7.2 刷新 Provider 容量

**POST /api/v1/quota/providers/{provider_id}/refresh**

## 7.3 Provider 容量状态

**GET /api/v1/quota/providers/{provider_id}**

---

# 八、系统配置 API

## 8.1 获取配置

**GET /api/v1/settings**

## 8.2 更新 OpenList 配置

**PUT /api/v1/settings/openlist**

## 8.3 更新 aria2 配置

**PUT /api/v1/settings/aria2**

## 8.4 下载策略

**PUT /api/v1/settings/download-policy**

## 8.5 Quota 策略

**PUT /api/v1/settings/quota-policy**

---

# 九、调试 API

## 9.1 下载调试

**GET /api/v1/debug/download/{task_id}**

## 9.2 Provider 调试

**GET /api/v1/debug/providers/{provider_id}**

## 9.3 继承关系

**GET /api/v1/debug/inheritance**

---

# 十、设计说明（重要）

## 10.1 架构核心思想

* OpenBridge = OpenList + Provider 扩展 + aria2 编排层
* 不替代 OpenList，仅做增强
* Provider = OpenList Storage 的“增强视图”

## 10.2 关键能力

* 二次认证（解决 quota 与下载认证不一致问题）
* 下载链路解析（302 + header 注入）
* aria2 任务映射
* quota 统一抽象

## 10.3 Debug 模式

关键接口支持：

* redirect_chain
* final_url
* headers
* cache_hit
* fallback_reason

---

# 十一、后续优化建议（开发阶段）

* [ ] 增加错误码规范文档
* [ ] 定义 provider_id 生成规则
* [ ] 增加 WebSocket（任务推送）
* [ ] 增加限流与权限模型
* [ ] OpenAPI / Swagger 自动生成

---

# 备注

该文档为临时开发版本，字段与接口可能调整。
