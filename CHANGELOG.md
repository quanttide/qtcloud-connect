# 更新日志

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)，
项目遵循 [语义化版本](https://semver.org/spec/v2.0.0.html) 规范。

> CLI scope 变更见 `src/cli/CHANGELOG.md`（对应 `cli/*` 版本线）。

## [0.0.2] - 2026-05-17

### Added

- Provider: FastAPI provider with storage, agents, API routes (38 单元测试)
- CLI: Typer CLI client 调用 Provider API（含集成测试）
- CLI: Provider API 子进程集成测试
- 聊天与会话生命周期集成测试
- PRD、IXD、用户指南等产品文档

### Changed

- CLI: 迁移至 Typer，API 路由从 /api/v1 改为 /api
- Provider: SQLite 数据库移至 data/ 目录，修复 SQLite 线程安全问题

### Fixed

- CLI: httpx 超时配置适配 LLM 请求

### Removed

- 废弃 docs/qa 文档

## [0.0.1] - 2026-01-04

### 前端

增加白板页面

### 产品需求文档

- 增加用户故事地图

### 交互设计文档

- 增加对话、白板、画布三个交互模式
- 增加白板页面组件设计
