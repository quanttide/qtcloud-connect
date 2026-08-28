# qtcloud-connect provider (Go)

量潮沟通云 — 人机沟通共识引擎 API (Go 实现)

## 运行

```bash
# 设置环境变量
export DB_PATH=data/qtcloud-connect.db
export PORT=8000

# 运行
go run main.go
```

## API

### 消息 API

- `GET /api/messages` - 列出所有消息
- `GET /api/messages/{id}` - 获取消息详情

### 共识 API

- `GET /api/consensuses` - 列出所有共识
- `POST /api/consensuses/confirm` - 确认共识
- `POST /api/consensuses/deprecate` - 废弃共识

### 健康检查

- `GET /healthz` - 健康检查

## 数据库

使用 SQLite 存储数据，数据库文件默认位于 `data/qtcloud-connect.db`。

## 开发

```bash
# 运行测试
go test ./...

# 构建
go build -o qtcloud-connect-provider main.go
```
