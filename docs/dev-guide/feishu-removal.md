# 飞书功能废弃与移除

## 结论

qtcloud-connect 的飞书能力分两步退出：v0.1.x 冻结并标记废弃，v0.2.0 起分批删除。删除的不是平台能力，而是本仓重复实现的那一层；替代在 qtcloud-feishu，逐项映射见下。

命令面与库分开处理。库与示例没有对外影响，替代判据过了就能删；`mail` 与 `notice` 是用户入口，必须等 qtcloud-feishu 有可用入口，或因迁移明确改走其他流程后再删。库能力对等不等于用户能操作。

## 判定与范围

同时满足两条的模块才算本次移除对象：直接调 lark-cli，或读飞书的身份、邮箱与返回结构；只服务飞书场景。

Provider 的消息、共识、关系与图 API，Studio 的共识页面，以及 `examples/connect-api` 都不在此列。根仓 `docs/gallery` 与 `docs/handbook` 里的 lark-cli 原始命令序列是案例素材，不随本仓移除。

## 替代映射

| 模块 | 对外形态 | 替代 | 阻塞条件 |
|------|----------|------|----------|
| `chat.rs` | 库，两个示例使用 | qtcloud-feishu `im_messages_search` 的 `query` 参数与「搜聊天记录」定义 | 搜索判据全绿 |
| `notice.rs` | `notice` 子命令 | `im_chat_search`、`contact_search_user`、`im_messages_send` 与两条定义 | 命令面可用 |
| `mail.rs` | `mail send`、`template`、`log` | `mail_send`、`mail_draft_create`、`mail_draft_send` 与操作日志 | 命令面可用 |
| `email.rs` | 库 | `mail_list`、`mail_messages` 与「读投递邮件」定义 | 多文件夹游标判据 |
| `report.rs` | 库与测试 | 行数据契约；排版与统计归领域仓（qtrecurit 或报表仓） | 领域侧对照测试 |
| `config.rs`、`cli_config.rs` | 库，读 profile 规则 | 领域仓（qtrecurit）保留规则与 profile 路径 | 同上 |
| `extract.rs` | 库，两个示例使用 | 搜索行与镜像；模型与提示词归领域仓 | 依赖树无 LLM |
| `examples/cli/qtclass.rs`、`examples/cli/qtrecurit.rs` | 示例 | 迁各自领域仓 | 领域仓接手 |
| `main.rs`、`lib.rs` | 子命令与导出 | 只留共识命令；摄取入口另议 | 上述完成 |
| `Cargo.toml` | 依赖 | 删 `quanttide-agent`、`regex`、`serde_yaml`；`chrono` 随 `mail.rs` 退出 | 模块与示例删完 |
| `README.md`、用户指南、开发指南 | 文档 | 删发送通道与模板章节，改平台方向 | 与代码同批 |

替代方案的细节在各篇里：[聊天搜索](../../../qtcloud-feishu/src/cli/docs/dev-guide/im-search.md)、[群通知](../../../qtcloud-feishu/src/cli/docs/dev-guide/im-notice.md)、[邮件发送](../../../qtcloud-feishu/src/cli/docs/dev-guide/mail-send.md)、[邮件投递统计](../../../qtcloud-feishu/src/cli/docs/dev-guide/mail-report.md)、[聊天知识提取](../../../qtcloud-feishu/src/cli/docs/dev-guide/im-extract.md)。

## 阶段一：冻结与标记

v0.1.x 不再给飞书能力增加功能，只修安全与数据错误。

在 `README.md`、用户指南与 `src/cli/CHANGELOG.md` 标注废弃，写清替代位置与移除版本。若发布 v0.1.1，给 `mail`、`notice` 子命令在运行时向标准错误打一行提示，给公开模块加 `#[deprecated]`；提示只指向文档，不承诺替代命令的可用时间。

冻结期至少跨一个发布周期。没有补丁发布需求时，文档与 CHANGELOG 的标记不能省。

## 阶段二：移除库与示例

前置是替代判据全部通过，且组织内检索没有其他消费者。第一批只动没有命令入口的部分，构建失败的范围可控。

1. 删除 `chat.rs`、`email.rs`、`extract.rs`、`report.rs`、`config.rs`、`cli_config.rs`，同步删 `lib.rs` 的导出。
2. 在领域仓接手后删除 `examples/cli/qtclass.rs` 与 `examples/cli/qtrecurit.rs`，避免能力断档。
3. 删除 `quanttide-agent`、`regex`、`serde_yaml` 依赖，更新锁文件。
4. 跑 `cargo fmt --check`、`cargo test --locked`、`cargo clippy --locked -- -A warnings`，再用检索确认没有残留符号。

这一批不涉及用户命令，回滚只需还原提交，不影响 Provider 数据与 Studio。

## 阶段三：移除命令面

前置是 qtcloud-feishu 有可用的发送与群通知入口，或迁移已明确改走其他流程。命令面未定形时，`mail` 与 `notice` 保持冻结，不先删。

1. `main.rs` 去掉 `Notice`、`Mail` 两个子命令，`lib.rs` 去掉对应导出；只留共识命令，摄取入口另议。
2. 删除 `mail.rs`、`notice.rs`；`README.md` 删发送通道、模板、硬性约束与环境变量章节。
3. 清理 `SEND_LOG_DIR` 的说明与示例，删除 `chrono` 依赖（`mail.rs` 是最后一个使用者）。
4. 版本号升到 v0.2.0（若与阶段二同批），CHANGELOG 写 Removed 与替代入口，release notes 标明破坏性变更；若命令面晚于 v0.2.0 退出，则单独占一个次版本。

命令面删除与库删除分成两个提交，回滚时互不牵连。

## 数据与环境

发送日志 `.quanttide/logs/send.log` 不迁移、不删除。旧文件留作审计，qtcloud-feishu 的操作日志是新格式，两者不合并；需要对照时人工读。

`SEND_LOG_DIR` 与 `DEEPSEEK_API_KEY` 随能力迁走，`QTRECURIT_PROFILE` 归 qtrecurit。connect 的环境变量表只保留 Provider 相关项。

Provider 的 SQLite 表不变，移除不动任何数据。镜像数据归 qtcloud-feishu 的 XDG 目录，connect 从 v0.2 起只读消费。

## 验收

- 替代判据全绿，链接见上表。
- 检索 `lark-cli`、`SEND_LOG_DIR`、`DEEPSEEK_API_KEY`、`QTRECURIT_PROFILE`：代码与用户文档无命中；CHANGELOG 的历史条目与 gallery/handbook 的案例允许保留。
- 检索 `qtcloud-connect mail` 与 `qtcloud-connect notice`：用户文档没有可执行指引。
- `cargo fmt --check`、`cargo test --locked`、`cargo clippy --locked -- -A warnings`、`cargo publish --dry-run` 通过。
- `go test ./...`、`go vet ./...`、`go build ./cmd/server` 通过；Studio 的 `flutter analyze`、`flutter test`、`flutter build web` 通过。
- `tests/verify-v0.1.ps1` 仍覆盖 Provider、CLI、Studio 的共识闭环。

## 回滚与发布

v0.1.x 的标签与 crates.io 版本保留，可回退。移除分两批提交，回滚只还原对应提交；没有数据库迁移，回滚不涉及数据修复。

发布顺序：qtcloud-feishu 先出版本并通过判据，qtcloud-connect 再出 v0.2.0。合并前检查组织内仓库与文档案例的引用，合并后更新根仓子模块指针。
