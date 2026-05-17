"""
集成测试 fixture：启动真实 provider 子进程。
"""

from __future__ import annotations

import os
import subprocess
import sys
import time
from pathlib import Path
from typing import Generator

import httpx
import pytest
from typer.testing import CliRunner

PROVIDER_DIR = Path(__file__).resolve().parent.parent.parent / "provider"


@pytest.fixture(scope="session")
def live_url() -> Generator[str, None, None]:
    """启动真实 provider 子进程，返回 base URL。"""
    # 使用 provider 的 Python（它拥有所有依赖）
    provider_python = str(PROVIDER_DIR / ".venv" / "bin" / "python3")
    if not os.path.exists(provider_python):
        provider_python = str(PROVIDER_DIR / ".venv" / "bin" / "python")
    if not os.path.exists(provider_python):
        pytest.fail(f"provider Python not found at {provider_python}")

    proc = subprocess.Popen(
        [
            provider_python,
            "-m",
            "uvicorn",
            "app:app",
            "--host",
            "127.0.0.1",
            "--port",
            "8765",
        ],
        cwd=str(PROVIDER_DIR),
        env=os.environ,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    # 等待 provider 就绪
    url = "http://127.0.0.1:8765"
    for _ in range(30):
        try:
            r = httpx.get(f"{url}/health", timeout=1)
            if r.is_success:
                break
        except Exception:
            time.sleep(0.5)
    else:
        proc.kill()
        pytest.fail("provider 未能启动")

    yield url

    proc.terminate()
    try:
        proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        proc.kill()


@pytest.fixture
def runner() -> CliRunner:
    return CliRunner()


@pytest.fixture
def cli_invoke(runner: CliRunner, live_url: str):
    """运行 CLI 命令，指向真实 provider。"""
    from app.main import app

    def _invoke(args: list[str]) -> tuple[int, str, str]:
        result = runner.invoke(app, args + ["--url", live_url])
        return result.exit_code, result.stdout, result.stderr

    return _invoke
