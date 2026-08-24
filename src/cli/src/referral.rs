use anyhow::{Context, Result};
use chrono::Local;
use clap::Args;
use std::path::PathBuf;

use qtcloud_connect_send::mail::{LarkMailer, SendLogEntry, append_send_log};

#[derive(Args)]
pub struct ReferralArgs {
    #[command(subcommand)]
    pub action: ReferralAction,
}

#[derive(clap::Subcommand)]
pub enum ReferralAction {
    /// 凭证化推荐：生成凭证号 → 草稿 → 确认 → 发送 → 写台账
    Send(ReferralSendArgs),
}

#[derive(Args)]
pub struct ReferralSendArgs {
    /// 候选人姓名
    #[arg(long)]
    pub name: String,

    /// 候选人邮箱
    #[arg(long)]
    pub candidate_email: String,

    /// 目标企业名
    #[arg(long)]
    pub company: String,

    /// 台账 CSV 路径（默认 $REFERRAL_CSV 或 data/referral/referrals.csv）
    #[arg(long)]
    pub csv: Option<String>,

    /// 确认后直接发送（默认只生成草稿）
    #[arg(long)]
    pub confirm_send: bool,

    /// 发送前打印将执行的命令，不执行
    #[arg(long)]
    pub dry_run: bool,
}

/// 生成凭证号：REF-YYYYMMDD-NNN
pub fn gen_referral_code(now: chrono::DateTime<chrono::Local>, seq: u32) -> String {
    format!("REF-{}-{:03}", now.format("%Y%m%d"), seq)
}

/// 推荐邮件正文（只给已验证事实 + 品格结论，不给判分细节）
pub fn build_referral_body(name: &str, company: &str, code: &str) -> String {
    format!(
        r#"您好，

我们向贵司（{company}）推荐候选人 {name}。

候选人经量潮招聘流程评估，材料真实，责任心评级为中等及以上，配合度良好。以下为已验证事实：
- 候选人基本信息、简历、问卷材料均与我们核实一致
- 评估过程无诚信问题

如需进一步了解候选人的具体情况，欢迎随时联系我们。

此邮件由量潮推荐系统自动生成，凭证号：{code}。

量潮科技 招聘团队"#
    )
}

/// 默认台账路径
pub fn default_csv_path() -> PathBuf {
    std::env::var("REFERRAL_CSV")
        .map(PathBuf::from)
        .unwrap_or_else(|_| PathBuf::from("data/referral/referrals.csv"))
}

/// 追加一行台账（fail-closed：写入失败不静默）
pub fn append_referral_csv(csv: Option<&str>, row: &[&str]) -> Result<()> {
    let path = match csv {
        Some(c) => PathBuf::from(c),
        None => default_csv_path(),
    };
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent).context("创建台账目录失败")?;
    }
    let line = row.join(",");
    use std::io::Write;
    let mut f = std::fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(&path)
        .context("打开台账失败")?;
    writeln!(f, "{}", line).context("写入台账失败")?;
    Ok(())
}

pub fn run(args: &ReferralArgs) {
    match &args.action {
        ReferralAction::Send(a) => cmd_send(a),
    }
}

fn cmd_send(args: &ReferralSendArgs) {
    let now = Local::now();
    let seq = (now.timestamp() % 1000) as u32;
    let code = gen_referral_code(now, seq);

    let subject = format!("人才推荐：{} → {}", args.name, args.company);
    let body = build_referral_body(&args.name, &args.company, &code);

    println!("=== 推荐凭证 ===");
    println!("凭证号: {}", code);
    println!("候选人: {} <{}>", args.name, args.candidate_email);
    println!("企业:   {}", args.company);
    println!();

    let mailer = LarkMailer;
    match mailer.send(
        &args.candidate_email,
        &subject,
        &body,
        None,
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
                "{} 凭证 {} | 状态: {}",
                if args.dry_run {
                    "[dry-run]"
                } else if sent {
                    "✓ 已发送"
                } else {
                    "✓ 已生成草稿"
                },
                code,
                status
            );

            // 回写台账（车越 bug 防复发：发送动作与台账联动）
            if !args.dry_run {
                let row = [
                    &code,
                    &now.format("%Y-%m-%d").to_string(),
                    &args.name,
                    &args.candidate_email,
                    &args.company,
                    "",
                    "",
                    &status,
                    "",
                ];
                if let Err(e) = append_referral_csv(args.csv.as_deref(), &row) {
                    eprintln!("错误: 台账写入失败（fail-closed，不会假装成功）: {e:#}");
                    std::process::exit(1);
                }
                println!("✓ 台账已更新");

                // 发送日志（只记元数据）
                let entry = SendLogEntry {
                    time: now.to_rfc3339(),
                    to: args.candidate_email.clone(),
                    subject: subject.clone(),
                    template: "referral".into(),
                    status,
                    draft_id: id,
                    note: Some(format!("凭证 {code}")),
                };
                if let Err(e) = append_send_log(None, &entry) {
                    eprintln!("警告: 发送日志写入失败: {e:#}");
                }
            }
        }
        Err(e) => {
            eprintln!("错误: {e:#}");
            std::process::exit(1);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_gen_referral_code_format() {
        let now = chrono::DateTime::parse_from_rfc3339("2026-08-22T10:00:00+08:00")
            .unwrap()
            .with_timezone(&chrono::Local);
        let code = gen_referral_code(now, 42);
        assert_eq!(code, "REF-20260822-042");
    }

    #[test]
    fn test_referral_code_uniqueness_with_seq() {
        let now = chrono::DateTime::parse_from_rfc3339("2026-08-22T10:00:00+08:00")
            .unwrap()
            .with_timezone(&chrono::Local);
        let a = gen_referral_code(now, 1);
        let b = gen_referral_code(now, 2);
        assert_ne!(a, b);
    }

    #[test]
    fn test_build_referral_body() {
        let body = build_referral_body("张三", "示例企业", "REF-20260822-001");
        assert!(body.contains("张三"));
        assert!(body.contains("示例企业"));
        assert!(body.contains("REF-20260822-001"));
        // 只给已验证事实 + 品格结论，不给判分细节
        assert!(!body.contains("51 分"));
        assert!(!body.contains("评分"));
        assert!(!body.contains("判分"));
        assert!(body.contains("材料真实"));
        assert!(body.contains("责任心评级"));
    }

    #[test]
    fn test_referral_csv_append() {
        let tmp = std::env::temp_dir().join(format!("referral-csv-{}", std::process::id()));
        std::fs::create_dir_all(&tmp).unwrap();
        let csv = tmp.join("referrals.csv");
        // 先写表头
        std::fs::write(
            &csv,
            "凭证号,日期,候选人,候选人邮箱,企业,岗位,推荐理由摘要,结果,企业反馈\n",
        )
        .unwrap();

        let row = [
            "REF-1",
            "2026-08-22",
            "张三",
            "wu@example.com",
            "示例企业",
            "",
            "",
            "draft",
            "",
        ];
        append_referral_csv(Some(csv.to_str().unwrap()), &row).unwrap();

        let content = std::fs::read_to_string(&csv).unwrap();
        assert!(content.contains("REF-1"));
        assert!(content.contains("wu@example.com"));
        assert_eq!(content.lines().count(), 2);

        std::fs::remove_dir_all(&tmp).ok();
    }

    #[test]
    fn test_default_csv_path() {
        // SAFETY: 测试环境单线程，无并发读环境变量
        unsafe { std::env::remove_var("REFERRAL_CSV") };
        assert_eq!(
            default_csv_path(),
            PathBuf::from("data/referral/referrals.csv")
        );
    }
}
