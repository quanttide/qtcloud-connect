"""
消息智能体（System 1 — 快思考）
"""

from __future__ import annotations

import requests

from app.config import settings

MESSAGE_SYSTEM_PROMPT = """你是 qtcloud-connect 的消息智能体（System 1）。

你的角色是与人类进行自然、流畅的对话。

## 核心职责
- 理解人类意图，生成自然回复
- 保持对话的连贯性与节奏感
- 在回复中自然地呼应已有共识

## 当前活跃共识
{consensus_summary}
"""


class MessageAgent:
    def __init__(self) -> None:
        self.api_key = settings.llm_api_key.get_secret_value()
        self.base_url = "https://api.deepseek.com"

    def get_consensus_summary(self, confirmed_consensuses: list[dict]) -> str:
        if not confirmed_consensuses:
            return "（暂无已确认的共识）"
        lines = ["## 已确认的共识"]
        for c in confirmed_consensuses:
            lines.append(f"- {c['content']}")
        return "\n".join(lines)

    def reply(
        self, user_message: str, history: list[dict], confirmed_consensuses: list[dict]
    ) -> str:
        consensus_summary = self.get_consensus_summary(confirmed_consensuses)
        system_prompt = MESSAGE_SYSTEM_PROMPT.format(
            consensus_summary=consensus_summary
        )

        messages = [{"role": "system", "content": system_prompt}]
        for h in history:
            messages.append(h)
        messages.append({"role": "user", "content": user_message})

        resp = requests.post(
            f"{self.base_url}/v1/chat/completions",
            headers={"Authorization": f"Bearer {self.api_key}"},
            json={"model": "deepseek-v4-flash", "messages": messages, "stream": False},
            timeout=60,
        )
        resp.raise_for_status()
        return resp.json()["choices"][0]["message"]["content"]
