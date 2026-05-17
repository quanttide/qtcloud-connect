"""集成测试：CLI ↔ 真实 provider API 完整交互。"""

from __future__ import annotations

import uuid


class TestCliIntegration:
    """所有命令测试真实 provider。"""

    def test_list_messages_empty(self, cli_invoke) -> None:
        """空消息列表。"""
        ec, out, err = cli_invoke(["messages"])
        assert ec == 0, f"exit={ec}, err={err!r}"

    def test_list_consensuses_empty(self, cli_invoke) -> None:
        """空共识列表。"""
        ec, out, err = cli_invoke(["consensuses"])
        assert ec == 0, f"exit={ec}, err={err!r}"

    def test_confirm_not_found(self, cli_invoke) -> None:
        """确认不存在的共识。"""
        ec, out, err = cli_invoke(["confirm", str(uuid.uuid4())])
        assert ec == 1
        assert "未找到" in err

    def test_deprecate_not_found(self, cli_invoke) -> None:
        """废弃不存在的共识。"""
        ec, out, err = cli_invoke(["deprecate", str(uuid.uuid4())])
        assert ec == 1
        assert "未找到" in err

    def test_help(self, cli_invoke) -> None:
        """帮助信息。"""
        ec, out, err = cli_invoke(["--help"])
        assert ec == 0
        assert "messages" in out
        assert "consensuses" in out
        assert "confirm" in out
        assert "deprecate" in out
