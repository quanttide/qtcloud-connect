# CHANGELOG

## [0.1.0-alpha.1] - 2026-08-24

### Added

- `mail` 招聘邮件封装（P1-P4）——模板化发送（草稿→确认→发送→日志）、模板查看、发送日志查询
- 模板机制（`MailTemplate` / `render_template` / `parse_vars`），机制保留、不写具体模板内容

### Changed

- 发送通道能力（LarkMailer / 模板机制 / 发送日志）保留在 CLI 内（曾抽独立 crate 后又撤销，通道能力回归 CLI）
- 包元数据补齐（description / license / repository），对齐 qtcloud-devops-cli 惯例

### Removed

- `referral` 子命令迁至招聘域 qtrecurit（凭证化人才推荐业务，issue #1）
- 招聘话术模板（referral/training/exam）随业务迁至 qtrecurit，`mail` 仅保留 raw 发送通道
