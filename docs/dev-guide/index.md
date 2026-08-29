# 开发者指南详细内容

## 当前目标

v0.1 的工程目标是让团队能真实记录并查看共识进展：

1. CLI 写入共识记录。
2. Provider 持久化共识记录。
3. Studio 读取 Provider 并展示共识追溯页面。

## 项目结构

```text
qtcloud-connect/
├── src/
│   ├── provider/   # Go Provider，SQLite 存储和 HTTP API
│   ├── cli/        # Rust CLI，邮件通道和 consensus 子命令
│   └── studio/     # Flutter Studio，共识追溯页面
├── docs/
├── examples/
└── tests/
```

## Provider

```bash
cd src/provider
go test ./...
go vet ./...
go build ./cmd/server
go run cmd/server/main.go
```

环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_PATH` | `data/qtcloud-connect.db` | SQLite 数据库路径 |
| `PORT` | `8000` | HTTP 端口 |
| `CONNECT_ALLOWED_ORIGINS` | 空 | 逗号分隔 CORS 白名单 |
| `CONNECT_AUTH_TOKEN` | 空 | 可选 Bearer token，配置后保护所有 `/api` 请求 |

Provider 默认允许目标 Studio 域名和本地开发端口跨域访问。`CONNECT_AUTH_TOKEN` 适合私有部署联调；公网部署仍必须通过网关补齐用户级认证、HTTPS 和限流。

## CLI

```bash
cd src/cli
cargo fmt --check
cargo test --locked
cargo clippy --locked -- -A warnings
cargo run -- consensus list
```

共识命令：

```bash
qtcloud-connect consensus create --title "共识标题" --description "共识描述"
qtcloud-connect consensus list
qtcloud-connect consensus show <consensus-id>
qtcloud-connect consensus update <consensus-id> --title "新标题" --description "新描述"
qtcloud-connect consensus confirm <consensus-id>
qtcloud-connect consensus deprecate <consensus-id>
```

## Studio

```bash
cd src/studio
flutter pub get
flutter analyze
flutter test
flutter build web --release --base-href /
flutter run -d chrome --dart-define=CONNECT_PROVIDER_ENDPOINT=http://localhost:8000/api
```

`CONNECT_PROVIDER_ENDPOINT` 默认值是 `http://localhost:8000/api`。部署到
`https://studio.connect.cloud.quanttide.com` 后，需要 Provider 的 CORS 白名单包含该域名。Studio 是静态 Web 产物，不能把 Provider 服务令牌传给浏览器；生产环境应由认证网关或反向代理完成用户认证，并在服务端持有 `CONNECT_AUTH_TOKEN`。

## 发布与部署

- CLI：发布前先走 `cli/v0.1.0-rc.x`，确认 CI、crates.io 发布和二进制 artifact 流程。
- Provider：先通过 `go test ./...`、`go vet ./...`、`go build ./cmd/server`，再部署到有持久化 `DB_PATH` 的运行环境。
- Studio：通过 `flutter build web --release --base-href /` 产出静态资源，部署到目标域名。
- 根仓库：子模块提交合并后再更新根仓库指针。
