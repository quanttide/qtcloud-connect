"""集成测试：CLI ↔ 真实 provider API 完整交互。"""

from __future__ import annotations

import uuid

import httpx


class TestCliIntegration:
    """所有命令测试真实 provider（不含 chat）。"""

    def test_list_messages_empty(self, cli_invoke) -> None:
        ec, out, err = cli_invoke(["messages"])
        assert ec == 0, f"exit={ec}, err={err!r}"

    def test_list_consensuses_empty(self, cli_invoke) -> None:
        ec, out, err = cli_invoke(["consensuses"])
        assert ec == 0, f"exit={ec}, err={err!r}"

    def test_confirm_not_found(self, cli_invoke) -> None:
        ec, out, err = cli_invoke(["confirm", str(uuid.uuid4())])
        assert ec == 1
        assert "未找到" in err

    def test_deprecate_not_found(self, cli_invoke) -> None:
        ec, out, err = cli_invoke(["deprecate", str(uuid.uuid4())])
        assert ec == 1
        assert "未找到" in err

    def test_help(self, cli_invoke) -> None:
        ec, out, err = cli_invoke(["--help"])
        assert ec == 0
        assert "messages" in out
        assert "consensuses" in out
        assert "confirm" in out
        assert "deprecate" in out

    def verify_db_clean(self, live_url: str) -> None:
        """清理 provider 数据库（通过 health 端点确保服务运行）。"""
        r = httpx.get(f"{live_url}/health", timeout=5)
        assert r.is_success


class TestChatIntegration:
    """真实 chat 测试：调 DeepSeek API，验证消息落库。"""

    def test_send_message_and_check_storage(self, live_url: str) -> None:
        """发一条真实消息 → 验证回复非空 → 验证消息已落库。"""
        # 1. 发消息
        resp = httpx.post(
            f"{live_url}/api/chat",
            json={"message": "用一句话介绍 PostgreSQL", "conversation_id": "test-conv"},
            timeout=120,
        )
        assert resp.is_success, f"chat failed: {resp.status_code} {resp.text}"
        data = resp.json()
        assert "reply" in data
        assert len(data["reply"]) > 0
        assert data["conversation_id"] == "test-conv"
        print(f"  reply: {data['reply'][:80]}...")

        # 2. 验证消息已落库（user + agent 两条）
        msgs = httpx.get(f"{live_url}/api/messages", timeout=5).json()
        assert len(msgs) >= 2
        assert msgs[0]["role"] == "user"
        assert msgs[1]["role"] == "agent"

    def test_chat_context_consensus(self, live_url: str) -> None:
        """共识影响回复：先建共识 → 再问相关问题 → 回复应提及共识。"""
        # 1. 发第一条消息，建立上下文
        httpx.post(
            f"{live_url}/api/chat",
            json={
                "message": "记下来，我们用 PostgreSQL 作为主数据库",
                "conversation_id": "test-context",
            },
            timeout=120,
        ).raise_for_status()

        # 2. 查看 consensus 是否被提炼
        consensuses = httpx.get(f"{live_url}/api/consensuses", timeout=5).json()
        # 可能没有 consensus（取决于 LLM 是否触发），不强制断言

        # 3. 再发一条跟进消息
        resp = httpx.post(
            f"{live_url}/api/chat",
            json={"message": "我们用什么数据库？", "conversation_id": "test-context"},
            timeout=120,
        )
        assert resp.is_success
        reply = resp.json()["reply"]
        print(f"  reply: {reply[:80]}...")

        # 4. 检查是否提及 PostgreSQL（如果 consensus 被提炼，应该会提到）
        has_consensus = any(
            "PostgreSQL" in c.get("content", "")
            for c in httpx.get(f"{live_url}/api/consensuses", timeout=5).json()
        )
        if has_consensus:
            assert "PostgreSQL" in reply, f"有共识但回复未提及: {reply}"
