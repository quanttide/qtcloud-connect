# Provider API 参考文档

本文档详细描述 qtcloud-connect FastAPI 后端服务的 API 接口。

## 基础信息

- **基础 URL**：`http://localhost:8000/api/v1`
- **认证方式**：Bearer Token (JWT)
- **数据格式**：JSON
- **字符编码**：UTF-8
- **API 版本**：v1

## 认证授权

### JWT Token
- **获取方式**：通过 `/auth/login` 端点获取
- **有效期**：24 小时
- **刷新**：通过 `/auth/refresh` 端点刷新

### 请求头格式
```
Authorization: Bearer <token>
Content-Type: application/json
```

## 主要端点

### 对话管理

#### 获取对话列表
```
GET /conversations
```

**查询参数**：
- `page`：页码（默认：1）
- `per_page`：每页数量（默认：20）
- `search`：搜索关键词
- `type`：对话类型（group/private）

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "title": "项目讨论",
        "type": "group",
        "participants": ["user1", "user2"],
        "created_at": "2026-08-28T10:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "per_page": 20
  }
}
```

#### 创建新对话
```
POST /conversations
```

**请求体**：
```json
{
  "title": "项目讨论",
  "participants": ["user1", "user2"],
  "type": "group"
}
```

**响应示例**：
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": 1,
    "title": "项目讨论",
    "type": "group",
    "participants": ["user1", "user2"],
    "created_at": "2026-08-28T10:00:00Z"
  }
}
```

#### 获取对话详情
```
GET /conversations/{id}
```

**路径参数**：
- `id`：对话 ID

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "title": "项目讨论",
    "type": "group",
    "participants": ["user1", "user2"],
    "messages_count": 25,
    "created_at": "2026-08-28T10:00:00Z",
    "updated_at": "2026-08-28T12:00:00Z"
  }
}
```

#### 更新对话信息
```
PUT /conversations/{id}
```

**路径参数**：
- `id`：对话 ID

**请求体**：
```json
{
  "title": "新标题",
  "participants": ["user1", "user2", "user3"]
}
```

#### 删除对话
```
DELETE /conversations/{id}
```

**路径参数**：
- `id`：对话 ID

**响应示例**：
```json
{
  "code": 200,
  "message": "删除成功"
}
```

### 消息管理

#### 获取对话消息列表
```
GET /conversations/{id}/messages
```

**路径参数**：
- `id`：对话 ID

**查询参数**：
- `page`：页码（默认：1）
- `per_page`：每页数量（默认：50）
- `before`：获取此时间之前的消息
- `after`：获取此时间之后的消息

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "content": "你好，这是测试消息",
        "type": "text",
        "sender": "user1",
        "created_at": "2026-08-28T10:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "per_page": 50
  }
}
```

#### 发送新消息
```
POST /conversations/{id}/messages
```

**路径参数**：
- `id`：对话 ID

**请求体**：
```json
{
  "content": "你好，这是测试消息",
  "type": "text",
  "attachments": []
}
```

**响应示例**：
```json
{
  "code": 201,
  "message": "发送成功",
  "data": {
    "id": 1,
    "content": "你好，这是测试消息",
    "type": "text",
    "sender": "user1",
    "created_at": "2026-08-28T10:00:00Z"
  }
}
```

#### 获取消息详情
```
GET /messages/{id}
```

**路径参数**：
- `id`：消息 ID

#### 更新消息
```
PUT /messages/{id}
```

**路径参数**：
- `id`：消息 ID

**请求体**：
```json
{
  "content": "更新后的消息内容"
}
```

#### 删除消息
```
DELETE /messages/{id}
```

**路径参数**：
- `id`：消息 ID

### 联系人管理

#### 获取联系人列表
```
GET /contacts
```

**查询参数**：
- `page`：页码（默认：1）
- `per_page`：每页数量（默认：20）
- `search`：搜索关键词
- `group`：联系人分组

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "name": "张三",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "group": "同事",
        "created_at": "2026-08-28T10:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "per_page": 20
  }
}
```

#### 创建新联系人
```
POST /contacts
```

**请求体**：
```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "group": "同事"
}
```

#### 获取联系人详情
```
GET /contacts/{id}
```

**路径参数**：
- `id`：联系人 ID

#### 更新联系人信息
```
PUT /contacts/{id}
```

**路径参数**：
- `id`：联系人 ID

**请求体**：
```json
{
  "name": "李四",
  "email": "lisi@example.com",
  "phone": "13900139000",
  "group": "朋友"
}
```

#### 删除联系人
```
DELETE /contacts/{id}
```

**路径参数**：
- `id`：联系人 ID

### 白板管理

#### 获取白板列表
```
GET /whiteboards
```

**查询参数**：
- `page`：页码（默认：1）
- `per_page`：每页数量（默认：20）
- `search`：搜索关键词

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "title": "项目规划白板",
        "content": {},
        "owner": "user1",
        "created_at": "2026-08-28T10:00:00Z"
      }
    ],
    "total": 30,
    "page": 1,
    "per_page": 20
  }
}
```

#### 创建新白板
```
POST /whiteboards
```

**请求体**：
```json
{
  "title": "项目规划白板",
  "content": {}
}
```

#### 获取白板详情
```
GET /whiteboards/{id}
```

**路径参数**：
- `id`：白板 ID

#### 更新白板内容
```
PUT /whiteboards/{id}
```

**路径参数**：
- `id`：白板 ID

**请求体**：
```json
{
  "title": "新标题",
  "content": {
    "nodes": [],
    "edges": []
  }
}
```

#### 删除白板
```
DELETE /whiteboards/{id}
```

**路径参数**：
- `id`：白板 ID

### 数据分析

#### 对话统计分析
```
GET /analytics/conversations
```

**查询参数**：
- `start_date`：开始日期
- `end_date`：结束日期
- `group_by`：分组方式（day/week/month）

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_conversations": 1000,
    "active_conversations": 150,
    "average_messages_per_conversation": 25,
    "daily_stats": [
      {
        "date": "2026-08-28",
        "count": 50
      }
    ]
  }
}
```

#### 消息统计分析
```
GET /analytics/messages
```

**查询参数**：
- `start_date`：开始日期
- `end_date`：结束日期
- `group_by`：分组方式（day/week/month）
- `type`：消息类型（text/image/file）

#### 用户活跃度分析
```
GET /analytics/users
```

**查询参数**：
- `start_date`：开始日期
- `end_date`：结束日期
- `user_id`：用户 ID（可选）

#### 报表生成
```
GET /analytics/reports
```

**查询参数**：
- `type`：报表类型（daily/weekly/monthly）
- `format`：输出格式（json/csv/pdf）

## 错误处理

### HTTP 状态码
- `200`：成功
- `201`：创建成功
- `400`：请求参数错误
- `401`：未授权
- `403`：禁止访问
- `404`：资源不存在
- `429`：请求过于频繁
- `500`：服务器内部错误

### 错误响应格式
```json
{
  "code": 400,
  "message": "错误描述",
  "details": "详细错误信息",
  "timestamp": "2026-08-28T10:00:00Z"
}
```

### 常见错误

#### 400 Bad Request
```json
{
  "code": 400,
  "message": "请求参数错误",
  "errors": [
    {
      "field": "title",
      "message": "标题不能为空"
    }
  ]
}
```

#### 401 Unauthorized
```json
{
  "code": 401,
  "message": "未授权，请先登录"
}
```

#### 403 Forbidden
```json
{
  "code": 403,
  "message": "权限不足，无法访问该资源"
}
```

#### 404 Not Found
```json
{
  "code": 404,
  "message": "资源不存在"
}
```

#### 429 Too Many Requests
```json
{
  "code": 429,
  "message": "请求过于频繁，请稍后再试",
  "retry_after": 60
}
```

## 限流策略

### 请求限制
- **匿名用户**：100 次/小时
- **普通用户**：1000 次/小时
- **管理员**：10000 次/小时

### 限流响应
```json
{
  "code": 429,
  "message": "请求过于频繁，请稍后再试",
  "retry_after": 60
}
```

## 版本控制

### API 版本
- **当前版本**：v1
- **版本策略**：URL 路径版本控制
- **弃用策略**：提前 6 个月通知

### 版本迁移
- 新版本发布后，旧版本保留 6 个月
- 提供迁移指南和工具
- 监控 API 使用情况

## 监控和日志

### 监控指标
- 请求响应时间
- 错误率
- 吞吐量
- 资源使用率

### 日志格式
```json
{
  "timestamp": "2026-08-28T10:00:00Z",
  "level": "INFO",
  "message": "API 请求",
  "method": "POST",
  "path": "/api/v1/conversations",
  "status": 200,
  "duration": 150,
  "user_id": "user123"
}
```

## 测试和调试

### 测试工具
- **Postman**：API 测试和文档
- **Swagger UI**：交互式 API 文档（访问 `/docs`）
- **curl**：命令行测试

### 调试技巧
1. 使用 `--verbose` 参数查看详细请求
2. 检查请求头和响应头
3. 验证 JSON 格式和编码
4. 查看服务器日志

### 示例请求

#### 使用 curl 测试
```bash
# 获取 Token
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'

# 使用 Token 获取对话列表
curl -X GET http://localhost:8000/api/v1/conversations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

## 常见问题

### Q: 如何获取 API Token？
A: 通过 `/auth/login` 端点，使用用户名和密码获取 JWT Token。

### Q: API 请求频率有限制吗？
A: 是的，不同角色有不同的请求限制，详见限流策略部分。

### Q: 如何处理分页？
A: 使用 `page` 和 `per_page` 参数进行分页，响应中包含分页信息。

### Q: 支持哪些数据格式？
A: 目前只支持 JSON 格式，请求和响应都使用 UTF-8 编码。

### Q: 如何处理批量操作？
A: 对于批量操作，建议使用批量 API 端点（如果提供）或循环调用单个 API。

## 更新日志

### v1.0.0 (2026-08-28)
- 初始版本发布
- 支持对话、消息、联系人、白板管理
- 实现 JWT 认证
- 提供完整的 API 文档

### v1.1.0 (计划中)
- 添加数据分析 API
- 实现 WebSocket 实时通信
- 优化分页和搜索功能
