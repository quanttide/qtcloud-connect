# Provider API 参考

当前 Provider 是 Go + SQLite 实现，默认监听 `http://localhost:8000`，API 前缀为 `/api`。

## 运行

```bash
cd src/provider
go run cmd/server/main.go
```

环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_PATH` | `data/qtcloud-connect.db` | SQLite 数据库路径，首次启动会自动创建父目录 |
| `PORT` | `8000` | HTTP 服务端口 |
| `CONNECT_ALLOWED_ORIGINS` | 空 | 逗号分隔 CORS 白名单；为空时允许目标 Studio 域名和本地开发端口 |

## 共识 API

### 创建共识

```http
POST /api/consensuses
Content-Type: application/json
```

```json
{
  "title": "Studio 优先上线共识页面",
  "description": "v0.1 先打通 CLI 写入、Provider 持久化、Studio 展示。"
}
```

响应状态码：`201 Created`。

```json
{
  "id": "7d7a20cfd7d3446eaf4b9c531f9e3c18",
  "title": "Studio 优先上线共识页面",
  "description": "v0.1 先打通 CLI 写入、Provider 持久化、Studio 展示。",
  "status": "proposed",
  "created_at": "2026-08-29T04:00:00Z",
  "updated_at": "2026-08-29T04:00:00Z"
}
```

### 列出共识

```http
GET /api/consensuses?page=1&page_size=20
```

`page_size` 默认值为 20，最大值为 100；超过最大值时按 100 处理。

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

### 获取共识详情

```http
GET /api/consensuses/{id}
```

找不到记录时返回 `404` 和 `{ "error": "not found" }`。

### 更新共识

```http
PUT /api/consensuses/{id}
Content-Type: application/json
```

```json
{
  "title": "更新后的标题",
  "description": "更新后的描述"
}
```

### 确认共识

```http
POST /api/consensuses/confirm
Content-Type: application/json
```

```json
{ "consensus_id": "7d7a20cfd7d3446eaf4b9c531f9e3c18" }
```

### 废弃共识

```http
POST /api/consensuses/deprecate
Content-Type: application/json
```

```json
{ "consensus_id": "7d7a20cfd7d3446eaf4b9c531f9e3c18" }
```

## 消息 API

当前消息接口是只读基础接口：

```http
GET /api/messages
GET /api/messages/{id}
```

## 健康检查

```http
GET /healthz
```

响应：

```json
{ "status": "ok" }
```

## 浏览器访问

默认 CORS 白名单：

- `https://studio.connect.cloud.quanttide.com`
- `http://localhost:*`
- `http://127.0.0.1:*`

生产部署如果改用其他 Studio 域名，需要设置 `CONNECT_ALLOWED_ORIGINS`，例如：

```bash
CONNECT_ALLOWED_ORIGINS=https://studio.connect.cloud.quanttide.com,https://preview.example.com
```

不在白名单中的带 `Origin` 预检请求返回 `403`。Provider v0.1 不内置认证和限流，公网部署必须放在认证、HTTPS、限流网关之后。
