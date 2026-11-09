# 开发者指南

本目录记录 qtcloud-connect 的开发方式。v0.2 起的目标形态是沟通域的共识层：Provider 持久化消息、共识与关系，CLI 操作共识，Studio 展示共识图；飞书适配在 qtcloud-feishu，本仓最终不直接调用平台。

## 文档结构

- [详细内容](index.md)：平台方向、边界、架构、当前目标与路线。
- [群与渠道管理（lite）](lite.md)：群资产卡片化、名称与职责描述、群成员与群间关系，渠道先手动管理、适配层后接。
- [飞书功能废弃与移除](feishu-removal.md)：替代映射、冻结、分批删除与验收。

## 技术栈

- Provider：Go、SQLite、标准库 `net/http`
- CLI：Rust、clap、ureq
- Studio：Flutter、Dart、package:http

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

公网部署前必须确认 Provider 位于认证、HTTPS、限流网关之后；v0.1 Provider 本身不内置用户级认证。
