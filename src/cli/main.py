"""
qtcloud-connect CLI — 通过 provider API 调用的人机沟通共识引擎。
"""

from __future__ import annotations

from typing import Optional

import httpx
import typer

app = typer.Typer()
BASE_URL = "http://127.0.0.1:8000/api"


@app.callback(invoke_without_command=True)
def main(
    ctx: typer.Context,
    url: str = typer.Option(BASE_URL, "--url", help="Provider API base URL"),
) -> None:
    """qtcloud-connect — 人机沟通共识引擎 CLI"""
    if ctx.invoked_subcommand is not None:
        return
    repl(url)


def repl(base_url: str) -> None:
    """交互式 REPL"""
    client = httpx.Client(base_url=base_url.rstrip("/"))

    typer.echo("qtcloud-connect CLI — 输入消息开始对话，输入 /help 查看命令")
    typer.echo("=" * 50)

    while True:
        try:
            raw = input(">>> ").strip()
        except (EOFError, KeyboardInterrupt):
            typer.echo("\n再见。")
            break

        if not raw:
            continue

        if raw.startswith("/"):
            cmd, *parts = raw[1:].split(maxsplit=1)
            cmd = cmd.lower()

            if cmd in ("quit", "exit"):
                typer.echo("再见。")
                break

            elif cmd == "help":
                typer.echo("""命令：
  /quit              退出
  /messages          查看所有消息
  /consensuses       查看所有共识
  /confirm <id>      确认共识
  /deprecate <id>    废弃共识
  /help              显示此帮助""")

            elif cmd == "messages":
                resp = client.get("/messages")
                if resp.is_success:
                    for m in resp.json():
                        typer.echo(f"  [{m['id'][:8]}] {m['role']}: {m['content'][:60]}")
                else:
                    typer.echo(f"请求失败: {resp.status_code}")

            elif cmd == "consensuses":
                resp = client.get("/consensuses")
                if resp.is_success:
                    for c in resp.json():
                        typer.echo(f"  [{c['id'][:8]}] {c['status']}: {c['content'][:60]}")
                        if c["related_message_ids"]:
                            typer.echo(f"         ↳ 消息: {', '.join(m[:8] for m in c['related_message_ids'])}")
                else:
                    typer.echo(f"请求失败: {resp.status_code}")

            elif cmd == "confirm" and parts:
                resp = client.post("/consensuses/confirm", json={"consensus_id": parts[0]})
                if resp.is_success:
                    data = resp.json()
                    typer.echo(f"已确认共识 [{data['id'][:8]}]")
                else:
                    typer.echo("未找到该共识")

            elif cmd == "deprecate" and parts:
                resp = client.post("/consensuses/deprecate", json={"consensus_id": parts[0]})
                if resp.is_success:
                    data = resp.json()
                    typer.echo(f"已废弃共识 [{data['id'][:8]}]")
                else:
                    typer.echo("未找到该共识")

            else:
                typer.echo(f"未知命令: {raw}")

            continue

        typer.echo("  [发送消息...]", nl=False)
        resp = client.post("/chat", json={"message": raw, "conversation_id": "default"})
        if not resp.is_success:
            typer.echo(f" 请求失败: {resp.status_code}")
            continue

        data = resp.json()
        typer.echo(" ✓")
        typer.echo(f"  {data['reply']}")


@app.command()
def messages(
    url: str = typer.Option(BASE_URL, "--url", help="Provider API base URL"),
) -> None:
    """查看所有消息"""
    client = httpx.Client(base_url=url.rstrip("/"))
    resp = client.get("/messages")
    if resp.is_success:
        for m in resp.json():
            typer.echo(f"  [{m['id'][:8]}] {m['role']}: {m['content'][:60]}")


@app.command()
def consensuses(
    url: str = typer.Option(BASE_URL, "--url", help="Provider API base URL"),
) -> None:
    """查看所有共识"""
    client = httpx.Client(base_url=url.rstrip("/"))
    resp = client.get("/consensuses")
    if resp.is_success:
        for c in resp.json():
            typer.echo(f"  [{c['id'][:8]}] {c['status']}: {c['content'][:60]}")


@app.command()
def confirm(
    consensus_id: str = typer.Argument(help="共识 ID"),
    url: str = typer.Option(BASE_URL, "--url", help="Provider API base URL"),
) -> None:
    """确认共识"""
    client = httpx.Client(base_url=url.rstrip("/"))
    resp = client.post("/consensuses/confirm", json={"consensus_id": consensus_id})
    if resp.is_success:
        typer.echo(f"已确认共识 [{consensus_id[:8]}]")
    else:
        typer.echo("未找到该共识", err=True)
        raise typer.Exit(1)


@app.command()
def deprecate(
    consensus_id: str = typer.Argument(help="共识 ID"),
    url: str = typer.Option(BASE_URL, "--url", help="Provider API base URL"),
) -> None:
    """废弃共识"""
    client = httpx.Client(base_url=url.rstrip("/"))
    resp = client.post("/consensuses/deprecate", json={"consensus_id": consensus_id})
    if resp.is_success:
        typer.echo(f"已废弃共识 [{consensus_id[:8]}]")
    else:
        typer.echo("未找到该共识", err=True)
        raise typer.Exit(1)


if __name__ == "__main__":
    app()
