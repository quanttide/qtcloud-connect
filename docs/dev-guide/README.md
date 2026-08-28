# 开发者指南

本目录包含 qtcloud-connect 的开发者文档，帮助开发人员理解系统架构、进行二次开发和维护。

## 文档结构

| 文档 | 说明 |
|------|------|
| [详细内容](index.md) | 技术栈、开发流程、项目结构、常用命令等详细内容 |
| [架构设计](architecture.md) | 系统整体架构、模块划分和设计原则 |
| [开发环境](setup.md) | 本地开发环境搭建和配置 |
| [编码规范](coding-standards.md) | 代码风格、命名规范和最佳实践 |
| [部署指南](deployment.md) | 系统部署、配置和运维 |
| [贡献指南](contributing.md) | 如何参与项目开发和贡献代码 |

## 适用对象

- **后端开发**：负责 Provider 服务开发的工程师
- **前端开发**：负责 Studio 工作台开发的工程师
- **CLI 开发**：负责命令行工具开发的工程师
- **DevOps**：负责系统部署和运维的工程师

## 核心功能

### 系统架构
qtcloud-connect 采用前后端分离架构，包含三个核心组件：
- **Provider**：FastAPI 后端服务，提供 RESTful API
- **Studio**：Flutter 前端工作台，提供用户界面
- **CLI**：Rust 命令行工具，提供自动化操作

### 开发环境
- **后端**：Python 3.12+、uv、FastAPI
- **前端**：Flutter 3.10+、Dart
- **CLI**：Rust 1.70+、Cargo

### 开发流程
1. **环境搭建**：克隆仓库，安装依赖
2. **本地开发**：启动服务，运行测试
3. **代码提交**：创建分支，提交更改
4. **代码审查**：创建 Pull Request，接受审查
5. **部署上线**：合并代码，部署到生产环境

## 快速开始

### 1. 克隆仓库
```bash
git clone https://github.com/quanttide/qtcloud-connect.git
cd qtcloud-connect
```

### 2. 安装依赖
```bash
# Provider
cd src/provider
uv sync

# CLI
cd src/cli
cargo build

# Studio
cd src/studio
flutter pub get
```

### 3. 启动开发
```bash
# 启动 Provider
cd src/provider
uvicorn app.main:app --reload

# 运行 CLI
cd src/cli
cargo run -- --help

# 运行 Studio
cd src/studio
flutter run
```

## 相关文档

- [架构设计](architecture.md)：了解系统整体架构
- [开发环境](setup.md)：详细的环境搭建指南
- [编码规范](coding-standards.md)：代码风格和最佳实践
- [部署指南](deployment.md)：系统部署和运维
- [贡献指南](contributing.md)：如何参与项目开发

## 获取帮助

- **GitHub Issues**：[项目问题追踪](https://github.com/quanttide/qtcloud-connect/issues)
- **技术讨论**：内部飞书群组
- **代码审查**：Pull Request 审查流程
