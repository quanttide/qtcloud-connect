# 更新日志

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)，
项目遵循 [语义化版本](https://semver.org/spec/v2.0.0.html) 规范。

> CLI scope 变更见 `src/cli/CHANGELOG.md`（对应 `cli/*` 版本线）。

## [Unreleased]

### Changed

- CLI 发布工作流现在会将三平台二进制上传到 GitHub Release，并自动标记预发布版本。

### Added

- Provider: 新增共识创建、列表、详情、更新、确认和废弃接口，并补充 SQLite 存储测试。
- Studio: 默认入口切换为共识追溯页面，支持从 Provider 读取共识列表。

### Changed

- 文档更新为当前 Go Provider、Rust CLI、Flutter Studio 的真实运行方式。

### Fixed

- Provider 首次启动时自动创建默认 SQLite 数据目录。
- Provider 共识列表分页参数过大时不再触发切片越界 panic。

## [0.0.3] - 2026-08-28

### Changed

- 重构文档目录结构：将 docs/ 目录从按文档类型分类改为按用户角色分类
  - 新增 user-guide/：用户指南，包含操作手册、功能说明、常见问题
  - 新增 dev-guide/：开发者指南，包含架构设计、开发环境、编码规范
  - 新增 api-references/：API 参考文档，包含接口文档、CLI 命令、数据模型
  - 移除旧目录：brd/、mrd/、prd/、ixd/、qa/、adr/、dev/、user/
- 更新文档导航和索引，提供按角色和按功能的快速导航
- 完善 API 参考文档，包含详细的接口说明和示例
- 更新 README.md 中的仓库目录结构说明

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
