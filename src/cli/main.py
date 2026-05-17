"""
qtcloud-connect CLI — 通过 provider API 调用的人机沟通共识引擎客户端。

运行：
    uv run python main.py
"""

from __future__ import annotations

import argparse
import shlex

import httpx

BASE_URL = "http://127.0.0.1:8000/api/v1"


def main(argv: list[str] | None = None) -> None:
    parser = argparse.ArgumentParser(description="qtcloud-connect CLI")
    parser.add_argument("--url", default=BASE_URL, help="Provider API base URL")
    args = parser.parse_args(argv)
    base = args.url.rstrip("/")

    client = httpx.Client(base_url=base)
    conversation_id = "default"

    print("qtcloud-connect CLI — 输入消息开始对话，输入 /help 查看命令")
    print("=" * 50)

    while True:
        try:
            raw = input(">>> ").strip()
        except (EOFError, KeyboardInterrupt):
            print("\n再见。")
            break

        if not raw:
            continue

        if raw.startswith("/"):
            parts = shlex.split(raw)
            cmd = parts[0].lower()

            if cmd in ("/quit", "/exit"):
                print("再见。")
                break

            elif cmd == "/help":
                print("""命令列表：
  /quit              退出
  /messages          查看所有消息
  /consensuses       查看所有共识
  /confirm <id>      确认共识
  /deprecate <id>    废弃共识
  /history           （本地暂不支持）
  /help              显示此帮助""")

            elif cmd == "/messages":
                resp = client.get("/messages")
                if resp.is_success:
                    for m in resp.json():
                        print(f"  [{m['id'][:8]}] {m['role']}: {m['content'][:60]}...")
                else:
                    print(f"请求失败: {resp.status_code}")

            elif cmd == "/consensuses":
                resp = client.get("/consensuses")
                if resp.is_success:
                    for c in resp.json():
                        msg_ids = ", ".join(c["related_message_ids"])
                        print(f"  [{c['id'][:8]}] {c['status']}: {c['content'][:60]}")
                        if msg_ids:
                            print(f"         ↳ 消息: {msg_ids}")
                else:
                    print(f"请求失败: {resp.status_code}")

            elif cmd == "/confirm" and len(parts) >= 2:
                resp = client.post("/consensuses/confirm", json={"consensus_id": parts[1]})
                if resp.is_success:
                    data = resp.json()
                    print(f"已确认共识 [{data['id'][:8]}]: {data['status']}")
                else:
                    print("未找到该共识")

            elif cmd == "/deprecate" and len(parts) >= 2:
                resp = client.post("/consensuses/deprecate", json={"consensus_id": parts[1]})
                if resp.is_success:
                    data = resp.json()
                    print(f"已废弃共识 [{data['id'][:8]}]: {data['status']}")
                else:
                    print("未找到该共识")

            else:
                print(f"未知命令: {cmd}")

            continue

        # 正常对话
        print("  [发送消息...]", end=" ", flush=True)
        resp = client.post("/chat", json={
            "message": raw,
            "conversation_id": conversation_id,
        })

        if not resp.is_success:
            print(f"请求失败: {resp.status_code}")
            continue

        data = resp.json()
        print("✓")
        print(f"  {data['reply']}")


if __name__ == "__main__":
    import sys
    main(argv=sys.argv[1:])
