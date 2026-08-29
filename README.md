# 量潮沟通云(`qtcloud-connect`)

## 仓库目录

```
qtcloud-connect/
├── docs/ # 文档目录
│   ├── README.md # 文档概览和导航
│   ├── user-guide/ # 用户指南：操作手册、功能说明、常见问题
│   ├── dev-guide/ # 开发者指南：架构设计、开发环境、编码规范
│   └── api-references/ # API 参考：接口文档、CLI 命令、数据模型
└── src/ # 源代码
    ├── provider/ # 后端服务提供者（Go）— 共识信息持久化和 API
    ├── cli/ # 命令行工具（Rust）— 邮件发送/模板/日志、共识手动更新
    └── studio/ # 前端工作台（Flutter）— 共识追溯页面和用户工作站
```

## 共识追溯闭环

v0.1 当前目标是打通 `Provider -> CLI -> Studio` 的最小闭环：

- Provider 使用 Go + SQLite 暴露 `/api/consensuses`，保存共识标题、描述、状态和时间戳。
- CLI 使用 `qtcloud-connect consensus` 创建、查看、更新、确认和废弃共识。
- Studio 默认打开共识追溯页面，从 Provider 读取记录并展示 Message -> Consensus -> Memo 的工作流。

### 本地启动

```bash
# Provider
cd src/provider
go run cmd/server/main.go

# CLI 写入一条真实共识
cd ../cli
cargo run -- consensus create \
  --title "Studio 优先上线共识页面" \
  --description "v0.1 先打通 CLI 写入、Provider 持久化、Studio 展示。"

# Studio Web
cd ../studio
flutter run -d chrome --dart-define=CONNECT_PROVIDER_ENDPOINT=http://localhost:8000/api
```

## CLI 使用

### 共识管理

```bash
# 创建共识
qtcloud-connect consensus create --title "共识标题" --description "共识描述"

# 列表和详情
qtcloud-connect consensus list
qtcloud-connect consensus show <consensus-id>

# 更新状态和内容
qtcloud-connect consensus update <consensus-id> --title "新标题" --description "新描述"
qtcloud-connect consensus confirm <consensus-id>
qtcloud-connect consensus deprecate <consensus-id>
```

`--endpoint` 可覆盖默认 Provider 地址 `http://localhost:8000/api`。

### 发送通道

> 前置：安装 lark-cli 并完成登录（需 mail scope 授权：`lark-cli auth login --scope "mail:user_mailbox.message:send"`）
>
> 边界：本 CLI 只承载**发送通道**（raw 正文 → 草稿 → 确认 → 发送 → 日志）。
> 招聘业务（凭证化推荐 referral、招聘话术模板 referral/training/exam）已迁至
> 招聘域 [qtrecurit](https://github.com/quanttide/qtrecurit) CLI（issue #1）。

### 命令一览

```bash
# 发送邮件（默认只生成草稿，人工确认后才发）
qtcloud-connect mail send --to x@example.com --subject "自定义主题" --body "自定义正文"
qtcloud-connect mail send --to x@example.com --subject "自定义主题" --body-file body.md

# 确认后直接发送（人工确认过内容后用）
qtcloud-connect mail send --to x@example.com --subject "主题" --body "正文" --confirm-send

# 模板管理（通道侧仅 raw；招聘话术见 qtrecurit）
qtcloud-connect mail template --list

# 发送日志（只记元数据：时间/收件人/主题/状态，不记正文）
qtcloud-connect mail log --tail 20
```

### 模板

| 模板名 | 用途 |
|--------|------|
| `raw` | 自定义正文（--subject + --body 或 --body-file），通道侧唯一模板 |

招聘话术模板（`referral` 内推 / `training` 实训邀请 / `exam` 考核说明）已随业务迁至 `qtrecurit mail send --template <name>`。

### 硬性约束

- **草稿确认制**：默认只生成草稿（隐私邮件禁止全自动直发），人工确认后才 `--confirm-send`
- **fail-closed**：发送日志写入失败显式报错，不假装成功
- **日志隐私**：只记元数据（时间/收件人/主题/状态），不记正文

### 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `SEND_LOG_DIR` | `.quanttide/logs` | 发送日志目录 |

