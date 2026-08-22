use anyhow::{Context, Result};
use chrono::Local;
use clap::{Args, Subcommand};
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

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
    /// 模板化发送邮件：渲染模板 → 生成草稿 → 人工确认 → 发送 → 回写状态
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

    /// 模板名：referral（内推）/ training（实训邀请）/ exam（考核说明）/ raw（自定义正文）
    #[arg(long, default_value = "raw")]
    pub template: String,

    /// 主题（raw 模板必填）
    #[arg(long)]
    pub subject: Option<String>,

    /// 正文（raw 模板必填；或用 --body-file）
    #[arg(long)]
    pub body: Option<String>,

    /// 正文文件路径（替代 --body）
    #[arg(long)]
    pub body_file: Option<String>,

    /// 附件路径（逗号分隔）
    #[arg(long)]
    pub attach: Option<String>,

    /// 模板变量：key=value（逗号分隔多个），如 name=张三,company=示例企业
    #[arg(long)]
    pub vars: Option<String>,

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
    /// 模板名：referral / training / exam
    #[arg(long)]
    pub name: Option<String>,

    /// 列出所有模板
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
// 模板（三套话术，源自招聘工作/回复问题清单/回复问题清单.md）
// ─────────────────────────────────────────────────────────────

#[derive(Debug, Clone)]
pub struct MailTemplate {
    pub name: &'static str,
    pub description: &'static str,
    pub subject: &'static str,
    pub body: &'static str,
}

pub const TEMPLATES: &[MailTemplate] = &[
    MailTemplate {
        name: "referral",
        description: "企业内推沟通话术：向候选人确认是否接受推荐",
        subject: "量潮人才推荐沟通",
        body: r#"你好，

我们认真查看了你的简历，认为你的经验和能力都很突出。考虑到目前量潮的规模还比较小，入职后可能无法完全匹配你的职业期待和发展空间，因此我们正在考虑将一些能力优秀的候选人推荐到更大的平台，帮助你们在更适合的岗位上发挥所长。今天想先听听你对此的想法和意愿。

我们此前与西安交通大学樊老师有合作基础，樊老师的一位学生也曾在量潮实习。目前该学生在另一家公司实习，并负责协助所在公司招聘实习生，使用其个人内推码。基于樊老师为该学生提供的担保，我们与樊老师及该学生之间形成了信任关系，因此正在共同建立以内部推荐为主的招聘渠道，用于向该公司推荐合适的实习生人选。

关于你后续的安排，我们目前是这样考虑的：如果你接受推荐，为了便于统一管理和对外沟通，我们会以"量潮课堂的学生"这一身份为你进行推荐；如果你想继续留在量潮，就保持现在的安排不变；如果你暂时不想接受推荐，也不打算留在量潮，也请直接告诉我们，我们会尊重你的个人决定，不会勉强。这样主要是为了让推荐流程更规范，也避免信息混乱，同时也尊重你自己的选择。"#,
    },
    MailTemplate {
        name: "training",
        description: "实训邀请沟通话术：邀请通过初筛的候选人加入实训基地",
        subject: "量潮实训基地邀请",
        body: r#"{{name}}你好，

感谢你完成量潮科技的准入问卷。经评估，你已通过初筛，正式受邀加入量潮实训基地。

实训基地是量潮科技对外招聘考核的组成部分。你将在这里通过完成真实的工作任务接受考核，以实际产出代替答卷。

请扫码加入实训基地群（见附件二维码），进群后修改昵称为「{{name}}-岗位意向」。

具体考核规则将在群内发布，请关注群公告和资料。

期待在基地见到你。

量潮科技 招聘团队"#,
    },
    MailTemplate {
        name: "exam",
        description: "招聘考核说明话术：邀请候选人直接参与招聘考核",
        subject: "量潮招聘考核邀请",
        body: r#"你好，

我们认真看了你此前提交的材料及招聘流程中的整体表现，认为你目前展现出的能力和潜力符合量潮进一步招聘考核的要求，因此想邀请你直接参与招聘考核，也想先听听你的想法和意愿。

量潮目前的人才选拔以实际成果为核心，考核标准是：在相对开放的环境中，自主发现并提出有价值的问题，通过自己的方式创造实际成果。我们的考核不会以固定题目为主，而是希望你真正创造一个东西，以过程和产出作为判断依据。如果你暂时不适合这种方式，也可以选择实训等其他培养路径，通过阶段化任务逐步积累能力。需要提前说明的是，通过招聘考核代表你达到了进入量潮团队的人才选拔标准，但最终是否进入团队，还要看届时公司的岗位和项目情况。如果暂时没有合适岗位，我们也会优先考虑让你进入长期实训，或保留后续合作的可能。

如果你愿意参与招聘考核，可以直接回复我们，确认意愿后，我们会与你沟通具体考核方式和下一步安排。如果你希望先通过实训参与量潮，或者暂时不打算继续任何后续安排，也可以直接告诉我们。

期待你的回复。"#,
    },
];

pub fn find_template(name: &str) -> Option<&'static MailTemplate> {
    TEMPLATES.iter().find(|t| t.name == name)
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
// 发送日志（P4：只记元数据，不记正文）
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
    let vars = parse_vars(args.vars.as_deref());

    let (subject, body) = if args.template == "raw" {
        let subject = args.subject.clone().unwrap_or_else(|| {
            eprintln!("错误: raw 模板需要 --subject");
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
                eprintln!("错误: raw 模板需要 --body 或 --body-file");
                std::process::exit(1);
            }
        };
        (subject, body)
    } else {
        let tpl = match find_template(&args.template) {
            Some(t) => t,
            None => {
                eprintln!(
                    "错误: 未知模板 '{}'。可用: referral / training / exam / raw",
                    args.template
                );
                std::process::exit(1);
            }
        };
        (tpl.subject.to_string(), render_template(tpl, &vars))
    };

    // training 模板默认附带群二维码（$TRAINING_QR_PATH，敏感内容不放进仓库）
    let attach = if args.attach.is_some() {
        args.attach.clone()
    } else if args.template == "training" {
        match std::env::var("TRAINING_QR_PATH") {
            Ok(p) => Some(p),
            Err(_) => {
                eprintln!(
                    "警告: training 模板需要群二维码附件，请设置 TRAINING_QR_PATH 环境变量指向二维码图片"
                );
                None
            }
        }
    } else {
        None
    };

    // 未确认时只生成草稿
    let mailer = LarkMailer;
    match mailer.send(
        &args.to,
        &subject,
        &body,
        attach.as_deref(),
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
                    template: args.template.clone(),
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
        println!("可用模板:");
        for t in TEMPLATES {
            println!("  {} — {}", t.name, t.description);
        }
        return;
    }
    let name = match &args.name {
        Some(n) => n,
        None => {
            eprintln!(
                "用法: qtcloud-connect mail template --list 或 --name <referral|training|exam>"
            );
            std::process::exit(1);
        }
    };
    let tpl = match find_template(name) {
        Some(t) => t,
        None => {
            eprintln!("未知模板: {name}");
            std::process::exit(1);
        }
    };
    println!(
        "=== {} ===\n主题: {}\n\n{}",
        tpl.name, tpl.subject, tpl.body
    );
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
    fn test_find_template_all_three() {
        assert!(find_template("referral").is_some());
        assert!(find_template("training").is_some());
        assert!(find_template("exam").is_some());
        assert!(find_template("unknown").is_none());
    }

    #[test]
    fn test_render_template_vars() {
        let tpl = find_template("training").unwrap();
        let rendered = render_template(tpl, &[("name".to_string(), "张三".to_string())]);
        assert!(rendered.contains("张三你好"));
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
    fn test_render_no_unresolved_vars_in_default_templates() {
        // training 模板含 {{name}} 占位符，渲染后必须被替换
        let tpl = find_template("training").unwrap();
        let rendered = render_template(tpl, &[("name".to_string(), "测试".to_string())]);
        assert!(
            !rendered.contains("{{"),
            "渲染后仍有未解析占位符: {:?}",
            rendered
        );
        // referral 和 exam 无占位符
        for t in [
            find_template("referral").unwrap(),
            find_template("exam").unwrap(),
        ] {
            assert!(
                !t.body.contains("{{"),
                "模板 {} 有占位符: {:?}",
                t.name,
                t.body
            );
        }
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
