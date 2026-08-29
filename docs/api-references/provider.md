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
| `CONNECT_AUTH_TOKEN` | 空 | 可选 Bearer token；配置后所有 `/api` 请求都必须带 `Authorization: Bearer <token>` |

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

已确认或已废弃的共识不能再编辑标题和描述，更新时返回 `409 Conflict`。

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

## 共识图 API

共识图是 DAG，用来保存共识之间的可扩展链路。关系类型是字符串，可使用“前置条件”“支持”“反对”“补充”或团队自定义语义。

```http
POST /api/consensus-graphs
GET /api/consensus-graphs
GET /api/consensus-graphs/{id}
PUT /api/consensus-graphs/{id}
POST /api/consensus-graphs/{id}/nodes
DELETE /api/consensus-graphs/{id}/nodes/{consensus_id}
POST /api/consensus-graphs/{id}/relations
POST /api/consensus-graphs/{id}/edges
DELETE /api/consensus-graphs/{id}/edges/{relation_id}
```

`POST /api/consensus-graphs/{id}/relations` 会原子创建关系并加入图；如果会形成环，返回 `409 Conflict` 且不会写入孤立关系。

全局关系接口仍可用于复用或管理关系：

```http
POST /api/consensus-relations
GET /api/consensus-relations/{id}
DELETE /api/consensus-relations/{id}
GET /api/consensuses/{id}/relations?direction=incoming|outgoing|all
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

不在白名单中的带 `Origin` 预检请求返回 `403`。`CONNECT_AUTH_TOKEN` 只提供服务级 Bearer 校验；公网部署仍必须放在用户级认证、HTTPS、限流网关之后。
