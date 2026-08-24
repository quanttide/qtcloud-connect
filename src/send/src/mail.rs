//! 邮件发送通道：lark-cli 封装、模板渲染机制、发送日志。
//!
//! 本模块只含通道机制，不含业务话术内容（TEMPLATES 常量与 find_template
//! 由调用方持有——招聘话术归招聘域 qtrecurit，通用 raw 归 CLI 层）。

use anyhow::{Context, Result};
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

// ─────────────────────────────────────────────────────────────
// 模板渲染机制（内容由调用方提供）
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
// lark-cli 执行封装
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
// 单元测试（机制层，不依赖 lark-cli）
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
            template: "referral".into(),
            status: "draft".into(),
            draft_id: "d1".into(),
            note: None,
        };
        let e2 = SendLogEntry {
            time: "2026-08-22T10:01:00+08:00".into(),
            to: "b@example.com".into(),
            subject: "测试2".into(),
            template: "exam".into(),
            status: "sent".into(),
            draft_id: "d2".into(),
            note: Some("凭证 REF-20260822-001".into()),
        };
        append_send_log(Some(log_file.to_str().unwrap()), &e1).unwrap();
        append_send_log(Some(log_file.to_str().unwrap()), &e2).unwrap();

        let entries = read_send_log(Some(log_file.to_str().unwrap()), 10).unwrap();
        assert_eq!(entries.len(), 2);
        assert_eq!(entries[0].to, "a@example.com");
        assert_eq!(entries[1].status, "sent");
        assert_eq!(entries[1].note.as_deref(), Some("凭证 REF-20260822-001"));

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
