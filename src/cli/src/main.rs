use clap::{Parser, Subcommand};
use qtcloud_connect_cli::{consensus, mail, notice};

#[derive(Subcommand)]
enum Commands {
    /// 共识：创建、查询、更新、确认和废弃
    Consensus(consensus::ConsensusArgs),
    /// 发送飞书群通知并 @ 指定成员
    Notice(notice::NoticeArgs),
    /// 邮件：发送/模板/日志（发送通道；招聘话术见 qtrecurit CLI）
    Mail(mail::MailArgs),
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
        Some(Commands::Consensus(args)) => consensus::run(args),
        Some(Commands::Notice(args)) => notice::run(args),
        Some(Commands::Mail(args)) => mail::run(args),
        None => {}
    }
}
