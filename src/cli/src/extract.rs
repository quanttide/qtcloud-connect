//! 知识提取能力：从聊天/文本中提取结构化知识（LLM 驱动）。
//!
//! 替代原 knowl::acquire 在 connect 场景中的角色：
//! 聊天记录（LarkChatFetcher 搜索）→ extract_from_text → 结构化 JSON。

use std::env;

use anyhow::Result;
use quanttide_agent::llm::{CompleteOptions, LLM};
use quanttide_agent::message::Message;

/// 调用 LLM 从文本中提取知识（JSON）。
///
/// 依赖 `DEEPSEEK_API_KEY` 环境变量。
pub fn extract_from_text(text: &str, system_prompt: &str) -> Result<serde_json::Value> {
    let api_key = env::var("DEEPSEEK_API_KEY")?;
    let llm = LLM::new("deepseek-chat", "https://api.deepseek.com", &api_key);

    let messages = vec![
        Message::new("system", system_prompt),
        Message::new("user", &format!("从以下聊天记录中提取知识：\n\n{}", text)),
    ];

    let options = CompleteOptions {
        response_format: Some(serde_json::json!({"type": "json_object"})),
        ..Default::default()
    };

    let resp = llm.complete(&messages, options)?;
    quanttide_agent::llm::parse_structured_output(&resp.content)
        .map_err(|e| anyhow::anyhow!("解析失败: {}", e))
}
