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

## CLI 使用（招聘邮件封装）

> 前置：安装 lark-cli 并完成登录（需 mail scope 授权：`lark-cli auth login --scope "mail:user_mailbox.message:send"`）

### 命令一览

```bash
# 发送邮件（默认只生成草稿，人工确认后才发）
qtcloud-connect mail send --to 候选人@example.com --template exam
qtcloud-connect mail send --to a@x.com,b@y.com --template training --vars name=张三
qtcloud-connect mail send --to x@example.com --template raw --subject "自定义主题" --body "自定义正文"

# 确认后直接发送（人工确认过内容后用）
qtcloud-connect mail send --to x@example.com --template exam --confirm-send

# 模板管理
qtcloud-connect mail template --list
qtcloud-connect mail template --name exam

# 发送日志（只记元数据：时间/收件人/主题/状态，不记正文）
qtcloud-connect mail log --tail 20

# 凭证化人才推荐（凭证号 → 草稿 → 确认 → 发送 → 写台账）
qtcloud-connect referral send --name 张三 --candidate-email wu@example.com --company 示例企业
qtcloud-connect referral send --name 张三 --candidate-email wu@example.com --company 示例企业 --confirm-send
```

### 模板

| 模板名 | 用途 | 变量 |
|--------|------|------|
| `referral` | 企业内推沟通话术 | 无 |
| `training` | 实训邀请沟通话术（含群二维码附件需手动加 --attach） | `{{name}}` |
| `exam` | 招聘考核说明话术 | 无 |
| `raw` | 自定义正文（需 --subject + --body 或 --body-file） | — |

### 硬性约束

- **草稿确认制**：默认只生成草稿（隐私邮件禁止全自动直发），人工确认后才 `--confirm-send`
- **防漏发（车越 bug）**：`referral send` 发送后自动写台账 `data/referral/referrals.csv` + 发送日志
- **fail-closed**：台账/日志写入失败显式报错，不假装成功
- **日志隐私**：只记元数据（时间/收件人/主题/凭证号/状态），不记正文

### 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `SEND_LOG_DIR` | `.quanttide/logs` | 发送日志目录 |
| `REFERRAL_CSV` | `data/referral/referrals.csv` | 推荐台账路径（建议设为 qtrecurit-private 下的绝对路径） |
| `TRAINING_QR_PATH` | 无 | 实训邀请模板的群二维码图片路径（敏感内容，不放进仓库） |

