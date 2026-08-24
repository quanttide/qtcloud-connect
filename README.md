# 量潮沟通云(`qtcloud-connect`)

## 仓库目录

```
qtcloud-connect/
├── docs/ # 所有非代码文档
│ ├── brd/ # 商务需求文档：定价方案、计费模型、商业化策略
│ ├── mrd/ # 市场需求文档：客户需求收集、竞品分析、需求优先级（由产品运营维护）
│ ├── prd/ # 产品需求文档：功能规格、用户故事、验收标准
│ ├── ixd/ # 交互设计文档：原型、流程、组件交互规则
│ ├── qa/ # 质量保障文档：测试用例、验收 checklist、合规验证
│ ├── adr/ # 架构决策记录：技术选型与演进路径
│ ├── dev/ # 开发者文档：API 参考、集成指南、本地开发说明
│ └── user/ # 用户文档：Studio 工作台操作手册与配置指南
└── src/ # 源代码
    ├── provider/ # 后端服务提供者（FastAPI）— 为本系统和其他系统提供沟通管理能力
    ├── cli/ # 命令行工具（Rust）— 邮件发送/模板/日志、人才推荐凭证化
    └── studio/ # 前端工作台（Flutter）— 用户的工作站
```

## CLI 使用（发送通道）

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

