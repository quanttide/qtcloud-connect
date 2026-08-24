use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

use anyhow::{Context, Result};
use chrono::Local;
use clap::{Args, Subcommand};

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
// 模板机制（关键机制，保留；具体模板内容由业务域各自维护）
// ─────────────────────────────────────────────────────────────

#[derive(Debug, Clone)]
pub struct MailTemplate {
    pub name: &'static str,
    pub description: &'static str,
    pub subject: &'static str,
    pub body: &'static str,
}

/// 模板变量替换：{{key}} → value
pub fn render_template(template: &MailTemplate, vars: &[(String, String)]) -> String {
    let mut body = template.body.to_string();
    for (k, v) in vars {
        body = body.replace(&format!("{{{{{}}}}}", k), v);
    }
    body
}

/// 解析 --vars "name=张三,company=示例企业" 为键值对列表
pub fn parse_vars(raw: Option<&str>) -> Vec<(String, String)> {
    let mut out = Vec::new();
    if let Some(raw) = raw {
        for pair in raw.split(',') {
            if let Some((k, v)) = pair.split_once('=') {
                out.push((k.trim().to_string(), v.trim().to_string()));
            }
        }
    }
    out
}

// ─────────────────────────────────────────────────────────────
// lark-cli 执行封装（发送通道）
// ─────────────────────────────────────────────────────────────

pub struct LarkMailer;

impl LarkMailer {
    /// 发送邮件。默认生成草稿；confirm_send 时确认后直接发送。
    /// 返回 (draft_id 或 message_id, 是否实际发送)。
    pub fn send(
        &self,
        to: &str,
        subject: &str,
        body: &str,
        attach: Option<&str>,
        confirm_send: bool,
        dry_run: bool,
    ) -> Result<(String, bool)> {
        let mut args = vec![
            "mail",
            "+send",
            "--to",
            to,
            "--subject",
            subject,
            "--body",
            body,
            "--mailbox",
            "hr@quanttide.com",
        ];

        if let Some(att) = attach {
            args.extend(["--attach", att]);
        }
        if confirm_send {
            args.push("--confirm-send");
        }
        args.extend(["--as", "user", "--format", "json"]);

        if dry_run {
            eprintln!("[dry-run] lark-cli {}", args.join(" "));
            return Ok(("dry-run".to_string(), false));
        }

        let output = run_lark_cli(&args, Duration::from_secs(30))?;
        let data: serde_json::Value =
            serde_json::from_slice(&output.stdout).context("lark-cli +send 返回数据格式异常")?;

        let id = data["data"]["draft_id"]
            .as_str()
            .or_else(|| data["data"]["message_id"].as_str())
            .unwrap_or("")
            .to_string();
        Ok((id, confirm_send))
    }

    /// 发送已存在的草稿（+draft-send）
    pub fn send_draft(&self, draft_id: &str, dry_run: bool) -> Result<()> {
        let args = [
            "mail",
            "+draft-send",
            "--draft-id",
            draft_id,
            "--as",
            "user",
            "--format",
            "json",
        ];
        if dry_run {
            eprintln!("[dry-run] lark-cli {}", args.join(" "));
            return Ok(());
        }
        let output = run_lark_cli(&args, Duration::from_secs(30))?;
        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            anyhow::bail!("草稿发送失败: {}", stderr.trim());
        }
        Ok(())
    }
}

fn run_lark_cli(args: &[&str], timeout: Duration) -> Result<std::process::Output> {
    let child = Command::new("lark-cli")
        .args(args)
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .context("无法启动 lark-cli，请确认已安装并完成登录")?;

    let (tx, rx) = mpsc::channel();
    thread::spawn(move || {
        let result = child.wait_with_output();
        let _ = tx.send(result);
    });

    let output = rx
        .recv_timeout(timeout)
        .map_err(|_| anyhow::anyhow!("lark-cli 请求超时，请检查网络连接或认证状态"))?
        .context("lark-cli 进程异常退出")?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        anyhow::bail!("lark-cli 执行失败: {}", stderr.trim());
    }
    Ok(output)
}

// ─────────────────────────────────────────────────────────────
// 发送日志（只记元数据，不记正文）
// ─────────────────────────────────────────────────────────────

#[derive(Debug, serde::Serialize, serde::Deserialize)]
pub struct SendLogEntry {
    pub time: String,
    pub to: String,
    pub subject: String,
    pub template: String,
    pub status: String,
    pub draft_id: String,
    pub note: Option<String>,
}

pub fn default_log_dir() -> PathBuf {
    std::env::var("SEND_LOG_DIR")
        .map(PathBuf::from)
        .unwrap_or_else(|_| PathBuf::from(".quanttide/logs"))
}

/// 追加一条发送日志（fail-closed：写入失败不静默）
pub fn append_send_log(log_file: Option<&str>, entry: &SendLogEntry) -> Result<()> {
    let path = match log_file {
        Some(f) => PathBuf::from(f),
        None => {
            let dir = default_log_dir();
            std::fs::create_dir_all(&dir).context("创建日志目录失败")?;
            dir.join("send.log")
        }
    };
    let line = serde_json::to_string(entry).context("序列化日志失败")?;
    use std::io::Write;
    let mut f = std::fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(&path)
        .context("打开发送日志失败")?;
    writeln!(f, "{}", line).context("写入发送日志失败")?;
    Ok(())
}

/// 读取最近 N 条发送日志
pub fn read_send_log(log_file: Option<&str>, tail: usize) -> Result<Vec<SendLogEntry>> {
    let path = match log_file {
        Some(f) => PathBuf::from(f),
        None => default_log_dir().join("send.log"),
    };
    let content = std::fs::read_to_string(&path).context("读取发送日志失败")?;
    let mut entries: Vec<SendLogEntry> = content
        .lines()
        .filter_map(|l| serde_json::from_str(l).ok())
        .collect();
    let start = entries.len().saturating_sub(tail);
    entries.drain(..start);
    Ok(entries)
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
        println!("招聘话术模板（referral/training/exam）在 qtrecurit CLI");
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
    fn test_render_template_vars() {
        let tpl = MailTemplate {
            name: "test",
            description: "",
            subject: "s",
            body: "{{name}}你好，欢迎 {{company}}",
        };
        let rendered = render_template(&tpl, &[("name".to_string(), "张三".to_string())]);
        assert!(rendered.contains("张三你好"));
        assert!(rendered.contains("欢迎 {{company}}"));
        assert!(!rendered.contains("{{name}}"));
    }

    #[test]
    fn test_parse_vars() {
        let vars = parse_vars(Some("name=张三,company=示例企业"));
        assert_eq!(vars.len(), 2);
        assert_eq!(vars[0], ("name".to_string(), "张三".to_string()));
        assert_eq!(vars[1], ("company".to_string(), "示例企业".to_string()));
    }

    #[test]
    fn test_parse_vars_empty() {
        assert!(parse_vars(None).is_empty());
        assert!(parse_vars(Some("noequalsign")).is_empty());
    }

    #[test]
    fn test_default_log_dir() {
        // SAFETY: 测试环境单线程，无并发读环境变量
        unsafe { std::env::remove_var("SEND_LOG_DIR") };
        assert_eq!(default_log_dir(), PathBuf::from(".quanttide/logs"));
    }

    #[test]
    fn test_send_log_roundtrip() {
        let tmp = std::env::temp_dir().join(format!("send-log-test-{}", std::process::id()));
        std::fs::create_dir_all(&tmp).unwrap();
        let log_file = tmp.join("send.log");

        let e1 = SendLogEntry {
            time: "2026-08-22T10:00:00+08:00".into(),
            to: "a@example.com".into(),
            subject: "测试".into(),
            template: "raw".into(),
            status: "draft".into(),
            draft_id: "d1".into(),
            note: None,
        };
        append_send_log(Some(log_file.to_str().unwrap()), &e1).unwrap();

        let entries = read_send_log(Some(log_file.to_str().unwrap()), 10).unwrap();
        assert_eq!(entries.len(), 1);
        assert_eq!(entries[0].to, "a@example.com");
        assert_eq!(entries[0].status, "draft");

        std::fs::remove_dir_all(&tmp).ok();
    }

    #[test]
    fn test_send_log_tail() {
        let tmp = std::env::temp_dir().join(format!("send-log-tail-{}", std::process::id()));
        std::fs::create_dir_all(&tmp).unwrap();
        let log_file = tmp.join("send.log");
        for i in 0..5 {
            let e = SendLogEntry {
                time: format!("t{i}"),
                to: format!("to{i}@example.com"),
                subject: "s".into(),
                template: "raw".into(),
                status: "draft".into(),
                draft_id: format!("d{i}"),
                note: None,
            };
            append_send_log(Some(log_file.to_str().unwrap()), &e).unwrap();
        }
        let entries = read_send_log(Some(log_file.to_str().unwrap()), 2).unwrap();
        assert_eq!(entries.len(), 2);
        assert_eq!(entries[0].to, "to3@example.com");
        std::fs::remove_dir_all(&tmp).ok();
    }
}
