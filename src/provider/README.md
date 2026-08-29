# qtcloud-connect provider (Go)

量潮沟通云 — 人机沟通共识引擎 API (Go 实现)

## 项目结构

```
provider/
├── cmd/
│   └── server/         # 主服务器
│       └── main.go
├── internal/
│   ├── domain/         # 领域模型
│   │   └── models.go
│   ├── handler/        # HTTP 处理器
│   │   ├── router.go
│   │   ├── messages.go
│   │   ├── consensuses.go
│   │   └── consensus_graphs.go
│   └── store/          # 存储层
│       └── storage.go
├── data/               # 数据目录
├── go.mod
├── go.sum
└── README.md
```

## 运行

```bash
# 设置环境变量
export DB_PATH=data/qtcloud-connect.db
export PORT=8000
# 可选：配置后所有 /api 请求必须带 Authorization: Bearer <token>
export CONNECT_AUTH_TOKEN=replace-with-secret

# 运行
go run cmd/server/main.go
```

## API

### 消息 API

- `GET /api/messages` - 列出所有消息
- `GET /api/messages/{id}` - 获取消息详情

### 共识 API

- `POST /api/consensuses` - 创建共识
- `GET /api/consensuses?page=1&page_size=20` - 分页列出共识（`page_size` 最大 100）
- `GET /api/consensuses/{id}` - 获取共识详情
- `PUT /api/consensuses/{id}` - 更新共识标题和描述
- `POST /api/consensuses/confirm` - 确认共识，body: `{ "consensus_id": "..." }`
- `POST /api/consensuses/deprecate` - 废弃共识，body: `{ "consensus_id": "..." }`

### 共识关系和图 API

- `POST /api/consensus-relations` - 创建共识之间的关系，body: `{ "from": "...", "to": "...", "relation_type": "支持" }`
- `GET /api/consensus-relations/{id}` - 获取关系详情
- `DELETE /api/consensus-relations/{id}` - 删除关系并清理图内引用
- `GET /api/consensuses/{id}/relations` - 获取共识的入边/出边，可使用 `direction=incoming|outgoing|all`
- `POST /api/consensus-graphs` - 创建图
- `GET /api/consensus-graphs` - 列出图
- `GET /api/consensus-graphs/{id}` - 获取完整图详情，节点按拓扑顺序返回
- `PUT /api/consensus-graphs/{id}` - 更新图名称和描述
- `POST /api/consensus-graphs/{id}/nodes` - 将已有共识加入图
- `DELETE /api/consensus-graphs/{id}/nodes/{consensus_id}` - 从图中移出共识
- `POST /api/consensus-graphs/{id}/relations` - 原子创建关系并加入图
- `POST /api/consensus-graphs/{id}/edges` - 将已有关系加入图
- `DELETE /api/consensus-graphs/{id}/edges/{relation_id}` - 从图中移出关系

图是 DAG。Provider 会拒绝自环、缺失节点和成环关系；从图中移出节点或边不会删除全局共识或关系。Studio 使用图内原子建边接口，成环失败不会留下孤立关系。

默认允许 `https://studio.connect.cloud.quanttide.com`、`http://localhost:*` 和
`http://127.0.0.1:*` 访问共识 API。生产环境可用 `CONNECT_ALLOWED_ORIGINS` 配置逗号分隔的允许来源。
不在白名单中的浏览器预检请求返回 `403`。`CONNECT_AUTH_TOKEN` 为空时 Provider 不内置认证；配置后 `/api` 请求必须携带 Bearer token。公网部署仍必须放在用户级认证、HTTPS 和限流网关之后。

### 健康检查

- `GET /healthz` - 健康检查

## 数据库

使用 SQLite 存储数据，数据库文件默认位于 `data/qtcloud-connect.db`。

## 开发

```bash
# 运行测试
go test ./...

# 静态检查
go vet ./...

# 构建
go build -o qtcloud-connect-provider cmd/server/main.go
```
