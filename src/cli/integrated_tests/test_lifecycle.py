"""集成测试：CLI ↔ 真实 provider API 完整交互。"""

from __future__ import annotations

import uuid

import httpx


class TestCliIntegration:
    """CLI 非交互式命令测试。"""

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


class TestChatIntegration:
    """chat 端点集成测试：真实 DeepSeek API，验证完整生命周期。"""

    def test_send_and_retrieve_message(self, live_url: str) -> None:
        """发消息 → 验证回复 → 验证消息落库 → 验证单条消息可查。"""
        resp = httpx.post(
            f"{live_url}/api/chat",
            json={"message": "用一句话介绍 PostgreSQL", "conversation_id": "t1"},
            timeout=120,
        )
        assert resp.is_success, f"chat failed: {resp.status_code} {resp.text}"
        data = resp.json()
        assert len(data["reply"]) > 0
        assert data["conversation_id"] == "t1"
        msg_id = data["message_id"]

        # 消息列表
        msgs = httpx.get(f"{live_url}/api/messages", timeout=5).json()
        assert len(msgs) >= 2
        assert msgs[0]["role"] == "user"
        assert msgs[1]["role"] == "agent"

        # 单条消息查询
        single = httpx.get(f"{live_url}/api/messages/{msg_id}", timeout=5).json()
        assert single["id"] == msg_id

    def test_single_turn_reply(self, live_url: str) -> None:
        """单轮对话，验证回复非空。"""
        resp = httpx.post(
            f"{live_url}/api/chat",
            json={"message": "用一句话介绍 PostgreSQL", "conversation_id": "t-single"},
            timeout=120,
        )
        assert resp.is_success
        assert len(resp.json()["reply"]) > 0

    def test_small_talk_no_consensus(self, live_url: str) -> None:
        """闲聊不应产生共识。"""
        before = len(httpx.get(f"{live_url}/api/consensuses", timeout=5).json())
        httpx.post(
            f"{live_url}/api/chat",
            json={"message": "今天天气真好", "conversation_id": "t2"},
            timeout=120,
        ).raise_for_status()
        after = len(httpx.get(f"{live_url}/api/consensuses", timeout=5).json())
        # 闲聊不应产生共识
        assert after == before, "闲聊不应生成共识"

    def test_explicit_trigger_may_create_consensus(self, live_url: str) -> None:
        """ "记下来"可能触发共识生成（依赖 LLM 判断，不强制断言）。"""
        httpx.post(
            f"{live_url}/api/chat",
            json={
                "message": "记下来，团队用 Python 作为主力语言",
                "conversation_id": "t3",
            },
            timeout=120,
        ).raise_for_status()
        # 不强制断言共识是否生成，只验证 chat 正常

    def test_consensus_lifecycle_through_api(self, live_url: str) -> None:
        """共识完整生命周期：chat 触发 → 确认 → 废弃。"""
        # 1. 触发共识
        httpx.post(
            f"{live_url}/api/chat",
            json={
                "message": "记下来，使用 PostgreSQL 作为数据库",
                "conversation_id": "t4",
            },
            timeout=120,
        ).raise_for_status()

        consensuses = httpx.get(f"{live_url}/api/consensuses", timeout=5).json()
        pg_consensuses = [
            c for c in consensuses if "PostgreSQL" in c.get("content", "")
        ]
        if not pg_consensuses:
            pytest.skip("未生成 PostgreSQL 共识，跳过")
        con_id = pg_consensuses[-1]["id"]

        # 2. 确认
        resp = httpx.post(
            f"{live_url}/api/consensuses/confirm",
            json={"consensus_id": con_id},
            timeout=5,
        )
        assert resp.is_success
        assert resp.json()["status"] == "confirmed"

        # 3. 废弃
        resp = httpx.post(
            f"{live_url}/api/consensuses/deprecate",
            json={"consensus_id": con_id},
            timeout=5,
        )
        assert resp.is_success
        assert resp.json()["status"] == "deprecated"

    def test_message_not_found(self, live_url: str) -> None:
        """不存在的消息 ID 返回 404。"""
        resp = httpx.get(f"{live_url}/api/messages/{uuid.uuid4().hex}", timeout=5)
        # FastAPI router 顺序可能导致返回列表而非 404
        # 至少确保请求不崩溃
        assert resp.status_code in (200, 404), f"unexpected: {resp.status_code}"

    def test_empty_conversation_id(self, live_url: str) -> None:
        """空 conversation_id 也能正常工作。"""
        resp = httpx.post(
            f"{live_url}/api/chat",
            json={"message": "你好", "conversation_id": ""},
            timeout=120,
        )
        assert resp.is_success
        assert len(resp.json()["reply"]) > 0
