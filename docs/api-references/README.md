# API 参考文档

本目录包含 qtcloud-connect 的 API 参考文档，详细描述系统提供的所有 API 接口。

## 文档结构

| 文档 | 说明 |
|------|------|
| [Provider API](provider-api.md) | FastAPI 后端服务 API 接口文档 |
| [CLI 命令参考](cli-reference.md) | 命令行工具使用参考 |
| [数据模型](data-models.md) | 系统数据模型和实体定义 |
| [认证授权](authentication.md) | 认证和授权机制说明 |

## 适用对象

- **前端开发**：需要调用后端 API 的开发人员
- **第三方集成**：需要与系统集成的外部开发者
- **测试人员**：需要测试 API 功能的 QA 工程师
- **运维人员**：需要监控和调试 API 的运维工程师

## API 概览

### Provider API

#### 基础信息
- **基础 URL**：`http://localhost:8000/api/v1`
- **认证方式**：Bearer Token (JWT)
- **数据格式**：JSON
- **字符编码**：UTF-8

#### 主要端点

##### 对话管理
```
GET    /conversations              # 获取对话列表
POST   /conversations              # 创建新对话
GET    /conversations/{id}         # 获取对话详情
PUT    /conversations/{id}         # 更新对话信息
DELETE /conversations/{id}         # 删除对话
```

##### 消息管理
```
GET    /conversations/{id}/messages    # 获取对话消息列表
POST   /conversations/{id}/messages    # 发送新消息
GET    /messages/{id}                  # 获取消息详情
PUT    /messages/{id}                  # 更新消息
DELETE /messages/{id}                  # 删除消息
```

##### 联系人管理
```
GET    /contacts                  # 获取联系人列表
POST   /contacts                  # 创建新联系人
GET    /contacts/{id}             # 获取联系人详情
PUT    /contacts/{id}             # 更新联系人信息
DELETE /contacts/{id}             # 删除联系人
```

##### 白板管理
```
GET    /whiteboards               # 获取白板列表
POST   /whiteboards               # 创建新白板
GET    /whiteboards/{id}          # 获取白板详情
PUT    /whiteboards/{id}          # 更新白板内容
DELETE /whiteboards/{id}          # 删除白板
```

##### 数据分析
```
GET    /analytics/conversations   # 对话统计分析
GET    /analytics/messages        # 消息统计分析
GET    /analytics/users           # 用户活跃度分析
GET    /analytics/reports         # 报表生成
```

#### 请求示例

##### 创建对话
```json
POST /api/v1/conversations
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "项目讨论",
  "participants": ["user1", "user2"],
  "type": "group"
}
```

##### 发送消息
```json
POST /api/v1/conversations/123/messages
Content-Type: application/json
Authorization: Bearer <token>

{
  "content": "你好，这是测试消息",
  "type": "text",
  "attachments": []
}
```

#### 响应示例

##### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 123,
    "title": "项目讨论",
    "created_at": "2026-08-28T10:00:00Z"
  }
}
```

##### 错误响应
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

### CLI 命令参考

#### 基本用法
```bash
qtcloud-connect <command> [options]
```

#### 主要命令

##### 邮件管理
```bash
# 发送邮件
qtcloud-connect mail send --to <email> --template <template>

# 查看邮件模板
qtcloud-connect mail template --list

# 查看发送日志
qtcloud-connect mail log --tail <number>
```

##### 推荐管理
```bash
# 发送推荐
qtcloud-connect referral send --name <name> --candidate-email <email> --company <company>

# 查看推荐记录
qtcloud-connect referral list
```

##### 配置管理
```bash
# 查看配置
qtcloud-connect config list

# 设置配置
qtcloud-connect config set <key> <value>

# 重置配置
qtcloud-connect config reset
```

## 认证授权

### JWT Token
- **获取方式**：通过 `/auth/login` 端点获取
- **有效期**：24 小时
- **刷新**：通过 `/auth/refresh` 端点刷新

### 权限模型
- **用户角色**：普通用户、管理员、超级管理员
- **权限控制**：基于角色的访问控制 (RBAC)
- **资源权限**：读取、写入、删除、管理

### 安全措施
- 所有 API 必须通过 HTTPS 访问
- 敏感操作需要二次验证
- API 请求频率限制
- 输入数据验证和清洗

## 错误处理

### HTTP 状态码
- `200`：成功
- `201`：创建成功
- `400`：请求参数错误
- `401`：未授权
- `403`：禁止访问
- `404`：资源不存在
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
- **Swagger UI**：交互式 API 文档
- **curl**：命令行测试

### 调试技巧
1. 使用 `--verbose` 参数查看详细请求
2. 检查请求头和响应头
3. 验证 JSON 格式和编码
4. 查看服务器日志

## 常见问题

### Q: 如何获取 API Token？
A: 通过 `/auth/login` 端点，使用用户名和密码获取 JWT Token。

### Q: API 请求频率有限制吗？
A: 是的，不同角色有不同的请求限制，详见限流策略部分。

### Q: 如何处理分页？
A: 使用 `page` 和 `per_page` 参数进行分页，响应中包含分页信息。

### Q: 支持哪些数据格式？
A: 目前只支持 JSON 格式，请求和响应都使用 UTF-8 编码。

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
