# 开发者指南详细内容

本文档包含 qtcloud-connect 开发者指南的详细内容，包括技术栈、开发流程、项目结构、常用命令等。

## 技术栈

### 后端 (Provider)
- **框架**：FastAPI (Python 3.12+)
- **数据库**：SQLite
- **认证**：JWT + OAuth2
- **依赖管理**：uv

### 前端 (Studio)
- **框架**：Flutter
- **平台**：Android、iOS、Web、Desktop
- **状态管理**：Riverpod
- **依赖管理**：pub

### CLI 工具
- **语言**：Rust
- **框架**：clap
- **依赖管理**：Cargo

## 开发流程

### 1. 环境准备
```bash
# 克隆仓库
git clone https://github.com/quanttide/qtcloud-connect.git

# 安装依赖
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

### 2. 本地开发
```bash
# 启动 Provider
cd src/provider
uvicorn app.main:app --reload

# 运行 CLI
cd src/cli
cargo run -- <command>

# 运行 Studio
cd src/studio
flutter run
```

### 3. 测试
```bash
# Provider 测试
cd src/provider
pytest

# CLI 测试
cd src/cli
cargo test

# Studio 测试
cd src/studio
flutter test
```

### 4. 提交代码
```bash
# 创建功能分支
git checkout -b feature/your-feature

# 提交更改
git commit -m "feat: add your feature"

# 推送分支
git push origin feature/your-feature

# 创建 Pull Request
```

## 项目结构

```
qtcloud-connect/
├── src/
│   ├── provider/          # FastAPI 后端服务
│   │   ├── app/           # 应用代码
│   │   ├── tests/         # 测试代码
│   │   └── pyproject.toml # Python 依赖配置
│   ├── cli/               # Rust 命令行工具
│   │   ├── src/           # 源代码
│   │   ├── examples/      # 示例代码
│   │   └── Cargo.toml     # Rust 依赖配置
│   └── studio/            # Flutter 前端工作台
│       ├── lib/           # Dart 代码
│       ├── test/          # 测试代码
│       └── pubspec.yaml   # Flutter 依赖配置
├── docs/                  # 项目文档
├── examples/              # 示例项目
└── tests/                 # 集成测试
```

## 常用命令

### Provider
```bash
# 安装依赖
uv sync

# 运行测试
pytest

# 启动服务
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# 数据库迁移
alembic upgrade head
```

### CLI
```bash
# 构建
cargo build --release

# 运行测试
cargo test

# 代码检查
cargo clippy

# 格式化
cargo fmt
```

### Studio
```bash
# 获取依赖
flutter pub get

# 运行测试
flutter test

# 构建 APK
flutter build apk

# 构建 iOS
flutter build ios
```

## 调试技巧

### Provider 调试
- 使用 `--reload` 参数实现热重载
- 查看日志：`logging.basicConfig(level=logging.DEBUG)`
- 使用 debugger：在 VS Code 中配置 launch.json

### CLI 调试
- 使用 `RUST_LOG=debug` 环境变量
- 使用 `cargo expand` 查看宏展开
- 使用 `cargo clippy` 进行代码检查

### Studio 调试
- 使用 Flutter Inspector
- 使用 `flutter run --debug` 模式
- 查看日志：`flutter logs`

## 性能优化

### Provider
- 使用连接池管理数据库连接
- 实现缓存机制（Redis）
- 异步处理耗时任务

### CLI
- 使用 `--release` 模式编译
- 实现并行处理
- 优化内存使用

### Studio
- 使用 `const` 构造函数
- 实现懒加载
- 优化图片资源

## 安全 considerations

- 所有 API 必须进行身份验证
- 敏感数据必须加密存储
- 使用 HTTPS 进行通信
- 定期更新依赖包

## 故障排除

### 常见问题
1. **依赖安装失败**：检查网络连接和版本兼容性
2. **测试失败**：查看详细错误日志
3. **构建失败**：检查编译器版本和依赖版本
4. **运行时错误**：检查环境变量和配置文件

### 获取帮助
- 查看项目 GitHub Issues
- 联系核心开发团队
- 参考官方文档
