# 开发者指南

本目录记录 qtcloud-connect 的当前开发方式。v0.1 以共识追溯闭环为核心：
Go Provider 持久化共识，Rust CLI 手动更新共识，Flutter Studio 展示共识页面。

## 文档结构

| 文档 | 说明 |
|------|------|
| [详细内容](index.md) | 技术栈、开发流程、项目结构、常用命令 |

## 技术栈

- **Provider**：Go、SQLite、标准库 `net/http`
- **CLI**：Rust、clap、ureq
- **Studio**：Flutter、Dart、package:http

## 快速开始

```bash
# Provider
cd src/provider
go test ./...
go run cmd/server/main.go

# CLI
cd ../cli
cargo test --locked
cargo run -- consensus list

# Studio
cd ../studio
flutter analyze
flutter test
flutter run -d chrome --dart-define=CONNECT_PROVIDER_ENDPOINT=http://localhost:8000/api
```

## 发布前检查

```bash
cd src/provider && go test ./... && go vet ./... && go build ./cmd/server
cd ../cli && cargo fmt --check && cargo test --locked && cargo clippy --locked -- -A warnings
cd ../studio && flutter analyze && flutter test && flutter build web --release --base-href /
```

v0.1 的跨组件验收可以直接运行：

```powershell
pwsh -NoProfile -File tests/verify-v0.1.ps1
```

详细验收范围见 [v0.1 验收记录](../acceptance/v0.1.md)。

公网部署前必须确认 Provider 位于认证、HTTPS、限流网关之后；v0.1 Provider 本身不内置认证。
