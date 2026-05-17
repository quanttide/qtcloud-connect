"""测试 CLI 命令。"""

from __future__ import annotations

from unittest.mock import patch

import httpx
from typer.testing import CliRunner

from app.main import app


class TestCliCommands:
    """测试非交互式命令。"""

    def test_messages_empty(self, runner: CliRunner) -> None:
        with patch("app.main.httpx.Client") as mock_cls:
            mock_client = mock_cls.return_value
            mock_client.get.return_value = httpx.Response(200, json=[])
            result = runner.invoke(app, ["messages", "--url", "http://test/api"])
            assert result.exit_code == 0

    def test_messages_list(self, runner: CliRunner) -> None:
        with patch("app.main.httpx.Client") as mock_cls:
            mock_client = mock_cls.return_value
            mock_client.get.return_value = httpx.Response(200, json=[
                {"id": "a" * 32, "content": "hello", "role": "user", "created_at": "", "updated_at": None},
            ])
            result = runner.invoke(app, ["messages", "--url", "http://test/api"])
            assert result.exit_code == 0
            assert "hello" in result.stdout

    def test_consensuses_list(self, runner: CliRunner) -> None:
        with patch("app.main.httpx.Client") as mock_cls:
            mock_client = mock_cls.return_value
            mock_client.get.return_value = httpx.Response(200, json=[
                {"id": "c" * 32, "content": "PostgreSQL", "status": "confirmed",
                 "created_at": "", "related_message_ids": []},
            ])
            result = runner.invoke(app, ["consensuses", "--url", "http://test/api"])
            assert result.exit_code == 0
            assert "PostgreSQL" in result.stdout

    def test_confirm(self, runner: CliRunner) -> None:
        with patch("app.main.httpx.Client") as mock_cls:
            mock_client = mock_cls.return_value
            mock_client.post.return_value = httpx.Response(200, json={"id": "test123", "status": "confirmed"})
            result = runner.invoke(app, ["confirm", "test123", "--url", "http://test/api"])
            assert result.exit_code == 0
            assert "已确认" in result.stdout

    def test_confirm_not_found(self, runner: CliRunner) -> None:
        with patch("app.main.httpx.Client") as mock_cls:
            mock_client = mock_cls.return_value
            mock_client.post.return_value = httpx.Response(404, json={"error": "not found"})
            result = runner.invoke(app, ["confirm", "xxx", "--url", "http://test/api"])
            assert result.exit_code == 1
            assert "未找到" in result.stderr

    def test_deprecate(self, runner: CliRunner) -> None:
        with patch("app.main.httpx.Client") as mock_cls:
            mock_client = mock_cls.return_value
            mock_client.post.return_value = httpx.Response(200, json={"id": "test123", "status": "deprecated"})
            result = runner.invoke(app, ["deprecate", "test123", "--url", "http://test/api"])
            assert result.exit_code == 0
            assert "已废弃" in result.stdout

    def test_deprecate_not_found(self, runner: CliRunner) -> None:
        with patch("app.main.httpx.Client") as mock_cls:
            mock_client = mock_cls.return_value
            mock_client.post.return_value = httpx.Response(404, json={"error": "not found"})
            result = runner.invoke(app, ["deprecate", "xxx", "--url", "http://test/api"])
            assert result.exit_code == 1
            assert "未找到" in result.stderr

    def test_help(self, runner: CliRunner) -> None:
        result = runner.invoke(app, ["--help"])
        assert result.exit_code == 0
        assert "messages" in result.stdout
        assert "consensuses" in result.stdout
        assert "confirm" in result.stdout
        assert "deprecate" in result.stdout
