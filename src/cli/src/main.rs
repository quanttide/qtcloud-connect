use clap::{Parser, Subcommand};
use qtcloud_connect_cli::{mail, notice, referral};

#[derive(Subcommand)]
enum Commands {
    /// 发送飞书群通知并 @ 指定成员
    Notice(notice::NoticeArgs),
    /// 邮件：发送/模板/日志（招聘邮件封装）
    Mail(mail::MailArgs),
    /// 凭证化人才推荐（referral send）
    Referral(referral::ReferralArgs),
}

#[derive(Parser)]
#[command(name = "qtcloud-connect", version, about = "QtCloud Connect CLI")]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

fn main() {
    let cli = Cli::parse();
    match &cli.command {
        Some(Commands::Notice(args)) => notice::run(args),
        Some(Commands::Mail(args)) => mail::run(args),
        Some(Commands::Referral(args)) => referral::run(args),
        None => {}
    }
}
