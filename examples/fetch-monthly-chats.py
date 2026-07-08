#!/usr/bin/env python3
"""
Example: Fetch chat messages for a given month via lark-cli.

Demonstrates:
- Incremental saving (safe to interrupt)
- Checkpoint/resume support
- NDJSON + metadata artifacts

Usage:
    python3 fetch-monthly-chats.py [--month YYYY-MM]

Requires:
    pip install quanttide-agent
    lark-cli installed and authenticated
"""

import json, os, subprocess, sys, argparse
from datetime import datetime, timezone, timedelta

PAGE_SIZE = 50


def main():
    parser = argparse.ArgumentParser(description="Fetch monthly chat records via lark-cli")
    parser.add_argument("--month", help="Month to fetch (YYYY-MM, default: current)")
    parser.add_argument("--output-prefix", default="chat-records", help="Output file prefix")
    args = parser.parse_args()

    now = datetime.now(timezone(timedelta(hours=8)))
    if args.month:
        month = args.month
        start = datetime.strptime(f"{month}-01T00:00:00+08:00", "%Y-%m-%dT%H:%M:%S%z")
    else:
        month = now.strftime("%Y-%m")
        start = now.replace(day=1, hour=0, minute=0, second=0, microsecond=0)

    if start.month == 12:
        end = start.replace(year=start.year + 1, month=1, day=1)
    else:
        end = start.replace(month=start.month + 1, day=1)

    fmt = "%Y-%m-%dT%H:%M:%S+08:00"
    start_str, end_str = start.strftime(fmt), end.strftime(fmt)
    prefix = args.output_prefix
    mp, dp, sp = f"{prefix}-{month}-meta.json", f"{prefix}-{month}.ndjson", f"{prefix}-{month}.json"

    meta = None
    try:
        with open(mp) as f:
            meta = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        pass

    if meta and meta.get("completed"):
        print(f"{month} already complete ({meta['total_messages']} messages).", file=sys.stderr)
        return

    if meta:
        print(f"Resuming from checkpoint ({meta['total_messages']} saved, page_token={meta.get('page_token')})", file=sys.stderr)
    else:
        meta = {"month": month, "start": start_str, "end": end_str, "page_token": None, "completed": False, "total_messages": 0}
        with open(mp, "w") as f:
            json.dump(meta, f, ensure_ascii=False)
        print(f"Fetching {start_str} -> {end_str} ...", file=sys.stderr)

    while not meta["completed"]:
        cmd = ["lark-cli", "im", "+messages-search", "--query", "",
               "--start", start_str, "--end", end_str, "--page-size", str(PAGE_SIZE), "--format", "json"]
        if meta["page_token"]:
            cmd.extend(["--page-token", meta["page_token"]])
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode != 0:
            print(f"Error (code {result.returncode}):\n{result.stderr}", file=sys.stderr)
            sys.exit(1)
        body = json.loads(result.stdout)
        if not body.get("ok"):
            print(f"API error: {body}", file=sys.stderr)
            sys.exit(1)
        messages = body.get("data", {}).get("messages", [])
        with open(dp, "a") as f:
            for m in messages:
                f.write(json.dumps(m, ensure_ascii=False) + "\n")
        meta["total_messages"] += len(messages)
        meta["page_token"] = body.get("data", {}).get("page_token")
        meta["completed"] = not body.get("data", {}).get("has_more", False)
        tmp = mp + ".tmp"
        with open(tmp, "w") as f:
            json.dump(meta, f, ensure_ascii=False)
        os.replace(tmp, mp)
        print(f"  +{len(messages):4d}  total: {meta['total_messages']:5d}  {'done' if meta['completed'] else 'more...'}", file=sys.stderr)

    # Summary
    msgs = []
    if os.path.exists(dp):
        with open(dp) as f:
            for line in f:
                line = line.strip()
                if line:
                    msgs.append(json.loads(line))
    chats = {}
    for m in msgs:
        cid = m.get("chat_id")
        if cid:
            name = m.get("chat_name") or m.get("chat_partner", {}).get("open_id", cid)
            chats.setdefault(cid, {"name": name, "count": 0})
            chats[cid]["count"] += 1
            if m.get("chat_name"):
                chats[cid]["name"] = m["chat_name"]
    summary = {"month": month, "query": {"start": start_str, "end": end_str}, "total_messages": len(msgs), "completed": True,
               "conversations": sorted([{"name": v["name"], "message_count": v["count"]} for v in chats.values()], key=lambda x: -x["message_count"])}
    with open(sp, "w") as f:
        json.dump(summary, f, ensure_ascii=False, indent=2)
    print(f"\nDone! {meta['total_messages']} messages\n  Data: {dp}\n  Meta: {mp}\n  Sum:  {sp}", file=sys.stderr)


if __name__ == "__main__":
    main()
