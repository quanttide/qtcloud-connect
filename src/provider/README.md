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
│   │   └── consensuses.go
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

默认允许 `https://studio.connect.cloud.quanttide.com`、`http://localhost:*` 和
`http://127.0.0.1:*` 访问共识 API。生产环境可用 `CONNECT_ALLOWED_ORIGINS` 配置逗号分隔的允许来源。
不在白名单中的浏览器预检请求返回 `403`。Provider 本身不提供认证，公网部署必须放在认证、HTTPS 和限流网关之后。

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
