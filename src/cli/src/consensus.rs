use anyhow::{Context, Result};
use clap::{Args, Subcommand};
use serde::{Deserialize, Serialize};

const DEFAULT_ENDPOINT: &str = "http://localhost:8000/api";

#[derive(Args)]
pub struct ConsensusArgs {
    #[command(subcommand)]
    pub command: ConsensusCommand,
}

#[derive(Subcommand)]
pub enum ConsensusCommand {
    /// 创建一条共识记录
    Create(CreateConsensusArgs),
    /// 列出共识记录
    List(ListConsensusArgs),
    /// 查看共识详情
    Show(ShowConsensusArgs),
    /// 更新共识标题和描述
    Update(UpdateConsensusArgs),
    /// 确认一条共识
    Confirm(StatusConsensusArgs),
    /// 废弃一条共识
    Deprecate(StatusConsensusArgs),
}

#[derive(Args)]
pub struct CreateConsensusArgs {
    /// Provider API endpoint，例如 http://localhost:8000/api
    #[arg(long, default_value = DEFAULT_ENDPOINT)]
    pub endpoint: String,
    /// 共识标题
    #[arg(long)]
    pub title: String,
    /// 共识描述
    #[arg(long)]
    pub description: Option<String>,
}

#[derive(Args)]
pub struct ListConsensusArgs {
    /// Provider API endpoint，例如 http://localhost:8000/api
    #[arg(long, default_value = DEFAULT_ENDPOINT)]
    pub endpoint: String,
    /// 页码
    #[arg(long, default_value_t = 1)]
    pub page: usize,
    /// 每页数量
    #[arg(long = "page-size", default_value_t = 20)]
    pub page_size: usize,
}

#[derive(Args)]
pub struct ShowConsensusArgs {
    /// Provider API endpoint，例如 http://localhost:8000/api
    #[arg(long, default_value = DEFAULT_ENDPOINT)]
    pub endpoint: String,
    /// 共识 ID
    pub id: String,
}

#[derive(Args)]
pub struct UpdateConsensusArgs {
    /// Provider API endpoint，例如 http://localhost:8000/api
    #[arg(long, default_value = DEFAULT_ENDPOINT)]
    pub endpoint: String,
    /// 共识 ID
    pub id: String,
    /// 共识标题
    #[arg(long)]
    pub title: String,
    /// 共识描述
    #[arg(long)]
    pub description: Option<String>,
}

#[derive(Args)]
pub struct StatusConsensusArgs {
    /// Provider API endpoint，例如 http://localhost:8000/api
    #[arg(long, default_value = DEFAULT_ENDPOINT)]
    pub endpoint: String,
    /// 共识 ID
    pub id: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct CreateConsensusRequest {
    title: String,
    description: String,
}

impl CreateConsensusRequest {
    pub fn new(title: String, description: Option<String>) -> Self {
        Self {
            title,
            description: description.unwrap_or_default(),
        }
    }
}

#[derive(Debug, Clone, Serialize)]
struct UpdateConsensusRequest {
    title: String,
    description: String,
}

#[derive(Debug, Deserialize, Serialize)]
struct StatusConsensusRequest {
    consensus_id: String,
}

pub fn run(args: &ConsensusArgs) {
    if let Err(err) = run_result(args) {
        eprintln!("错误: {err:#}");
        std::process::exit(1);
    }
}

fn run_result(args: &ConsensusArgs) -> Result<()> {
    let client = ureq::agent();
    let value = match &args.command {
        ConsensusCommand::Create(args) => post_json(
            &client,
            endpoint_url(&args.endpoint, "/consensuses"),
            &CreateConsensusRequest::new(args.title.clone(), args.description.clone()),
        ),
        ConsensusCommand::List(args) => {
            let path = format!(
                "/consensuses?page={}&page_size={}",
                args.page, args.page_size
            );
            get_json(&client, endpoint_url(&args.endpoint, &path))
        }
        ConsensusCommand::Show(args) => get_json(
            &client,
            endpoint_url(&args.endpoint, &format!("/consensuses/{}", args.id)),
        ),
        ConsensusCommand::Update(args) => put_json(
            &client,
            endpoint_url(&args.endpoint, &format!("/consensuses/{}", args.id)),
            &UpdateConsensusRequest {
                title: args.title.clone(),
                description: match args.description.as_deref() {
                    Some(description) => description.to_string(),
                    None => {
                        let current = get_json(
                            &client,
                            endpoint_url(&args.endpoint, &format!("/consensuses/{}", args.id)),
                        )?;
                        merge_update_description(&current, None)?
                    }
                },
            },
        ),
        ConsensusCommand::Confirm(args) => post_json(
            &client,
            endpoint_url(&args.endpoint, "/consensuses/confirm"),
            &StatusConsensusRequest {
                consensus_id: args.id.clone(),
            },
        ),
        ConsensusCommand::Deprecate(args) => post_json(
            &client,
            endpoint_url(&args.endpoint, "/consensuses/deprecate"),
            &StatusConsensusRequest {
                consensus_id: args.id.clone(),
            },
        ),
    }?;

    println!(
        "{}",
        serde_json::to_string_pretty(&value).context("序列化响应失败")?
    );
    Ok(())
}

fn merge_update_description(
    current: &serde_json::Value,
    requested: Option<&str>,
) -> Result<String> {
    if let Some(description) = requested {
        return Ok(description.to_string());
    }

    current
        .get("description")
        .and_then(serde_json::Value::as_str)
        .map(str::to_string)
        .ok_or_else(|| anyhow::anyhow!("Provider 响应缺少 description 字段"))
}

fn get_json(client: &ureq::Agent, url: String) -> Result<serde_json::Value> {
    handle_response(with_auth(client.get(&url)).call())
}

fn post_json<T: Serialize>(
    client: &ureq::Agent,
    url: String,
    body: &T,
) -> Result<serde_json::Value> {
    handle_response(with_auth(client.post(&url)).send_json(serde_json::to_value(body)?))
}

fn put_json<T: Serialize>(
    client: &ureq::Agent,
    url: String,
    body: &T,
) -> Result<serde_json::Value> {
    handle_response(with_auth(client.put(&url)).send_json(serde_json::to_value(body)?))
}

fn with_auth(request: ureq::Request) -> ureq::Request {
    let token = std::env::var("CONNECT_AUTH_TOKEN").unwrap_or_default();
    let token = token.trim();
    if token.is_empty() {
        request
    } else {
        request.set("Authorization", &format!("Bearer {token}"))
    }
}

fn handle_response(
    response: std::result::Result<ureq::Response, ureq::Error>,
) -> Result<serde_json::Value> {
    match response {
        Ok(response) => response.into_json().context("Provider 响应不是合法 JSON"),
        Err(ureq::Error::Status(status, response)) => {
            let text = response.into_string().unwrap_or_default();
            anyhow::bail!("Provider 返回 {status}: {text}")
        }
        Err(err) => Err(err).context("Provider 请求失败"),
    }
}

pub fn endpoint_url(endpoint: &str, path: &str) -> String {
    format!(
        "{}/{}",
        endpoint.trim_end_matches('/'),
        path.trim_start_matches('/')
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use clap::Parser;

    #[derive(Parser)]
    struct TestCli {
        #[command(subcommand)]
        command: ConsensusCommand,
    }

    #[test]
    fn create_request_uses_spec_fields() {
        let request = CreateConsensusRequest::new(
            "Studio 优先上线共识页面".to_string(),
            Some("v0.1 先打通可见闭环".to_string()),
        );

        let value = serde_json::to_value(request).unwrap();
        assert_eq!(value["title"], "Studio 优先上线共识页面");
        assert_eq!(value["description"], "v0.1 先打通可见闭环");
        assert!(value.get("content").is_none());
    }

    #[test]
    fn endpoint_url_trims_duplicate_slashes() {
        assert_eq!(
            endpoint_url("http://localhost:8000/api/", "/consensuses"),
            "http://localhost:8000/api/consensuses"
        );
    }

    #[test]
    fn update_description_preserves_existing_value_when_not_requested() {
        let current = serde_json::json!({
            "id": "consensus-1",
            "title": "旧标题",
            "description": "原有描述"
        });

        let description = merge_update_description(&current, None).unwrap();

        assert_eq!(description, "原有描述");
    }

    #[test]
    fn parses_create_command() {
        let args = TestCli::try_parse_from([
            "consensus",
            "create",
            "--title",
            "共识页面上线",
            "--description",
            "Studio 可访问共识追溯图",
            "--endpoint",
            "http://localhost:8000/api",
        ])
        .unwrap();

        match args.command {
            ConsensusCommand::Create(create) => {
                assert_eq!(create.title, "共识页面上线");
                assert_eq!(
                    create.description.as_deref(),
                    Some("Studio 可访问共识追溯图")
                );
                assert_eq!(create.endpoint, "http://localhost:8000/api");
            }
            _ => panic!("expected create command"),
        }
    }
}
