"""
共识智能体（System 2 — 慢思考）
"""

from __future__ import annotations

import json
import re

from quanttide_agent import LLM
from quanttide_connect.models import ConsensusStatus, Message
from quanttide_connect.services.consensus import ConsensusService
from quanttide_connect.services.relation import RelationService

from app.config import settings
from app.storage import Storage

CONSENSUS_SYSTEM_PROMPT = """你是 qtcloud-connect 的共识智能体（System 2 — 慢思考）。

你的角色是观察对话，提炼共识，维护共识卡片库。你**不与人类直接对话**。

## 输出格式
当你判断需要提炼或更新共识时，输出以下格式的指令：

```
[CONSENSUS_ACTION]
action: propose | confirm | deprecate
content: 共识的具体内容
related_messages: ["消息ID1", "消息ID2"]
[/CONSENSUS_ACTION]
```

- `propose` — 从若干消息中提炼一个新的共识（提议状态）
- `confirm` — 确认一个已有提议的共识
- `deprecate` — 废弃一个不再适用的共识
"""


class ConsensusAgent:
    def __init__(
        self,
        storage: Storage,
        consensus_svc: ConsensusService,
        relation_svc: RelationService,
    ) -> None:
        self.storage = storage
        self.consensus_svc = consensus_svc
        self.relation_svc = relation_svc
        self._llm = LLM(
            model="deepseek-v4-flash",
            base_url="https://api.deepseek.com",
            api_key=settings.llm_api_key.get_secret_value(),
        )

    def observe(
        self, user_message: Message, agent_message: Message, history: list[dict]
    ) -> None:
        confirmed = self.storage.list_consensuses(ConsensusStatus.confirmed)
        proposed = self.storage.list_consensuses(ConsensusStatus.proposed)

        ctx = f"## 本轮用户消息\n{user_message.content}\n\n## 本轮 AI 回复\n{agent_message.content}\n\n"
        if confirmed:
            ctx += (
                "## 已确认的共识\n"
                + "\n".join(f"- {c.content}" for c in confirmed)
                + "\n\n"
            )
        if proposed:
            ctx += (
                "## 待确认的共识（proposed）\n"
                + "\n".join(f"- {c.content} (id: {c.id})" for c in proposed)
                + "\n\n"
            )
        ctx += "基于以上对话，如果有必要，输出共识操作指令。如果不需要，输出 [NO_ACTION]。\n"

        messages = [
            {"role": "system", "content": CONSENSUS_SYSTEM_PROMPT},
            *history[-6:],
            {"role": "user", "content": ctx},
        ]

        resp = self._llm.chat(messages)
        self._handle_instructions(resp.content)

    def _handle_instructions(self, output: str) -> None:
        if "[NO_ACTION]" in output or "[/CONSENSUS_ACTION]" not in output:
            return
        for block in re.findall(
            r"\[CONSENSUS_ACTION\](.*?)\[/CONSENSUS_ACTION\]", output, re.DOTALL
        ):
            action = self._parse_action(block)
            if action:
                self._execute(action)

    def _parse_action(self, block: str) -> dict | None:
        action: dict = {}
        for line in block.strip().splitlines():
            line = line.strip()
            if line.startswith("action:"):
                action["action"] = line.split(":", 1)[1].strip()
            elif line.startswith("content:"):
                action["content"] = line.split(":", 1)[1].strip()
            elif line.startswith("related_messages:"):
                raw = line.split(":", 1)[1].strip()
                try:
                    action["related_messages"] = json.loads(raw)
                except json.JSONDecodeError:
                    action["related_messages"] = re.findall(r'"([^"]+)"', raw)
        return action if "action" in action else None

    def _execute(self, action: dict) -> None:
        act = action.get("action")
        content = action.get("content", "")
        related = action.get("related_messages", [])

        if act == "propose" and content:
            self.consensus_svc.propose(content, related)
        elif act == "confirm":
            for pc in self.storage.list_consensuses(ConsensusStatus.proposed):
                if content and content in pc.content:
                    self.consensus_svc.confirm(pc.id)
                    return
        elif act == "deprecate":
            for cc in self.storage.list_consensuses(ConsensusStatus.confirmed):
                if content and content in cc.content:
                    self.consensus_svc.deprecate(cc.id)
                    return
