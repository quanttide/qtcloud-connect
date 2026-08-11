mod chat;
mod cli_config;
mod config;
mod email;
mod notice;
mod report;

use clap::{Parser, Subcommand};

#[derive(Subcommand)]
enum Commands {
    /// 发送飞书群通知并 @ 指定成员
    Notice(notice::NoticeArgs),
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
        None => {}
    }
}
