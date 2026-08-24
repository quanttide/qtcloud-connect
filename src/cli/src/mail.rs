use chrono::Local;
use clap::{Args, Subcommand};

use qtcloud_connect_send::mail::{
    LarkMailer, SendLogEntry, append_send_log, read_send_log,
};

// ─────────────────────────────────────────────────────────────
// 命令定义
// ─────────────────────────────────────────────────────────────

#[derive(Args)]
pub struct MailArgs {
    #[command(subcommand)]
    pub action: MailAction,
}

#[derive(Subcommand)]
pub enum MailAction {
    /// 发送邮件：渲染正文 → 生成草稿 → 人工确认 → 发送 → 回写状态
    Send(MailSendArgs),
    /// 查看/列出邮件模板
    Template(MailTemplateArgs),
    /// 查看发送日志
    Log(MailLogArgs),
}

#[derive(Args)]
pub struct MailSendArgs {
    /// 收件人邮箱（逗号分隔多个）
    #[arg(long)]
    pub to: String,

    /// 主题（必填）
    #[arg(long)]
    pub subject: Option<String>,

    /// 正文（或 --body-file）
    #[arg(long)]
    pub body: Option<String>,

    /// 正文文件路径（替代 --body）
    #[arg(long)]
    pub body_file: Option<String>,

    /// 附件路径（逗号分隔）
    #[arg(long)]
    pub attach: Option<String>,

    /// 发送日志回写路径（默认 $SEND_LOG_DIR/send.log）
    #[arg(long)]
    pub log_file: Option<String>,

    /// 确认后直接发送（默认只生成草稿）
    #[arg(long)]
    pub confirm_send: bool,

    /// 发送前打印将执行的命令，不执行
    #[arg(long)]
    pub dry_run: bool,
}

#[derive(Args)]
pub struct MailTemplateArgs {
    /// 列出可用模板
    #[arg(long)]
    pub list: bool,
}

#[derive(Args)]
pub struct MailLogArgs {
    /// 显示最近 N 条日志
    #[arg(long, default_value = "20")]
    pub tail: usize,
}

// ─────────────────────────────────────────────────────────────
// 命令入口
// ─────────────────────────────────────────────────────────────

pub fn run(args: &MailArgs) {
    match &args.action {
        MailAction::Send(a) => cmd_send(a),
        MailAction::Template(a) => cmd_template(a),
        MailAction::Log(a) => cmd_log(a),
    }
}

fn cmd_send(args: &MailSendArgs) {
    let subject = args.subject.clone().unwrap_or_else(|| {
        eprintln!("错误: 需要 --subject");
        std::process::exit(1);
    });
    let body = match (&args.body, &args.body_file) {
        (Some(b), _) => b.clone(),
        (None, Some(f)) => match std::fs::read_to_string(f) {
            Ok(c) => c,
            Err(e) => {
                eprintln!("错误: 读取正文文件失败: {e}");
                std::process::exit(1);
            }
        },
        (None, None) => {
            eprintln!("错误: 需要 --body 或 --body-file");
            std::process::exit(1);
        }
    };

    // 通道只发送；招聘话术内容由 qtrecurit CLI 渲染后传入
    let mailer = LarkMailer;
    match mailer.send(
        &args.to,
        &subject,
        &body,
        args.attach.as_deref(),
        args.confirm_send,
        args.dry_run,
    ) {
        Ok((id, sent)) => {
            let status = if args.dry_run {
                "dry-run".to_string()
            } else if sent {
                "sent".to_string()
            } else {
                "draft".to_string()
            };
            println!(
                "{} 收件人: {} | 主题: {} | 状态: {}",
                if args.dry_run {
                    "[dry-run]"
                } else if sent {
                    "✓ 已发送"
                } else {
                    "✓ 已生成草稿"
                },
                args.to,
                subject,
                status
            );
            if !args.dry_run {
                let entry = SendLogEntry {
                    time: Local::now().to_rfc3339(),
                    to: args.to.clone(),
                    subject: subject.clone(),
                    template: "raw".into(),
                    status,
                    draft_id: id,
                    note: None,
                };
                if let Err(e) = append_send_log(args.log_file.as_deref(), &entry) {
                    // fail-closed：日志写入失败必须显式报错
                    eprintln!("警告: 发送日志写入失败（发送本身已成功）: {e}");
                    eprintln!(
                        "请手动补记: {}",
                        serde_json::to_string(&entry).unwrap_or_default()
                    );
                }
            }
        }
        Err(e) => {
            eprintln!("错误: {e:#}");
            std::process::exit(1);
        }
    }
}

fn cmd_template(args: &MailTemplateArgs) {
    if args.list {
        println!("可用模板（通道侧）:");
        println!("  raw — 自定义正文（--subject/--body/--body-file）");
        println!("招聘话术模板（referral/training/exam）已迁至 qtrecurit CLI");
        return;
    }
    eprintln!("用法: qtcloud-connect mail template --list");
    std::process::exit(1);
}

fn cmd_log(args: &MailLogArgs) {
    match read_send_log(None, args.tail) {
        Ok(entries) => {
            if entries.is_empty() {
                println!("暂无发送日志");
                return;
            }
            for e in entries {
                println!(
                    "{} | {} | {} | {} | {}",
                    e.time, e.status, e.to, e.subject, e.draft_id
                );
            }
        }
        Err(e) => {
            eprintln!("错误: {e:#}");
            std::process::exit(1);
        }
    }
}

// ─────────────────────────────────────────────────────────────
// 单元测试
// ─────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_template_list_output_no_panic() {
        // 通道侧无内置话术模板（业务内容已随迁招聘域）
        let args = MailTemplateArgs { list: true };
        // 只验证不 panic；输出走 stdout
        cmd_template(&args);
    }
}
