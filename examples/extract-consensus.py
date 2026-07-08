#!/usr/bin/env python3
"""
Example: Extract organizational consensus from chat records using LLM.

Pipeline: load NDJSON → LLM extracts consensus per chat → merge & dedup

Outputs a JSON list of consensus items with content, type (constitutional/transactional),
category, participants, and source chat.

Usage:
    python3 extract-consensus.py --month 2026-07 --founder-name "张果" \
        --data-dir ./data

Requires:
    pip install quanttide-agent
    LLM API key configured
    Chat records in NDJSON format (one JSON object per line, grouped by chat_id)
"""

import argparse, asyncio, json, os, re, sys
from collections import defaultdict
from datetime import datetime
from pathlib import Path

from quanttide_agent.llm import AsyncLLM
from quanttide_agent.message import Message

CONCURRENCY = 5


def build_system_prompt(founder_name):
    return f"""你是一个沟通分析师，正在分析公司创始人「{founder_name}」的聊天记录。

## 背景
- 数据源是创始人「{founder_name}」的聊天记录，你只能看到{founder_name}参与或可见的对话
- 创始人是「{founder_name}」（聊天记录中 sender.name == "{founder_name}"）
- 「团队」= 除{founder_name}以外的成员

## 任务
阅读以下聊天记录片段，从中提取出**团队达成的共识**。

### 什么是"共识"
共识 = 团队在讨论后形成的**共同认可的决定、结论、约定或认知**。例如：
- "就用 Dropbox，先买一个月"（采购决策）
- "周五之前把方案发给客户"（行动约定）
- "这个需求放下一期做"（优先级排序）

### 什么不算"共识"
- ❌ 纯粹的事实陈述（"客户预算 30 万"）
- ❌ 个人的想法/建议（"我觉得可以试试"——未得到回应或认可）
- ❌ 闲聊（"今天天气不错"）
- ❌ 仅信息同步（"附件是合同草案"——无后续认可）

### 提取要求
每条共识需包含以下信息：
1. **content**: 共识内容（一句话，简洁完整）
2. **type**: `"constitutional"`（宪法级：组织原则、价值观、权限划分、长期策略；半年后仍有参考价值）或 `"transactional"`（事务级：具体分工、会议时间、采购、候选人处理；半年后大概率已失效）
3. **category**: 分类标签 — `"制度规范"` | `"任务分配"` | `"业务决策"` | `"观点确认"`
4. **participants**: 参与达成共识的人（去重，从消息 sender 提取姓名，去除 open_id 格式的 ID）
5. **source_chat**: 会话名称（群名或对方姓名）
6. **timestamp**: 共识形成时间（取相关消息中最后一条的时间）

## 输出格式
只返回 JSON 数组，不要其他文字：
[{{"content": "...", "type": "constitutional|transactional", "category": "制度规范|任务分配|业务决策|观点确认", "participants": ["..."], "source_chat": "...", "timestamp": "..."}}]

如果没有提取到任何共识，返回 []"""


def render_messages(messages):
    lines = []
    for m in messages:
        ts = m.get("create_time", "")
        sender = m.get("sender", {})
        name = sender.get("name", sender.get("id", "?")[:12])
        content = (m.get("content") or "")[:600]
        lines.append(f"[{ts}] {name}: {content}")
    return "\n".join(lines)


async def analyze_one(llm, system_prompt, chat_id, chat_label, messages):
    chat_text = render_messages(messages)
    time_start = messages[0].get("create_time", "")
    time_end = messages[-1].get("create_time", "")
    prompt = f"会话：{chat_label}\n时间：{time_start} ~ {time_end}\n\n{chat_text}"
    resp = await llm.complete(
        [
            Message(role="system", content=system_prompt),
            Message(role="user", content=prompt),
        ],
        model="deepseek-v4-flash", temperature=0.1, max_tokens=2000,
    )
    return {"chat_id": chat_id, "chat_label": chat_label,
            "time_start": time_start, "time_end": time_end,
            "llm_output": resp.content}


def parse_consensus_from_output(llm_output):
    raw = llm_output.get("llm_output", "")
    try:
        items = json.loads(raw)
    except json.JSONDecodeError:
        m = re.search(r"\[.*?\]", raw, re.DOTALL)
        if m:
            try:
                items = json.loads(m.group())
            except json.JSONDecodeError:
                items = []
        else:
            items = []
    if not isinstance(items, list):
        return []
    return [
        {
            "content": item.get("content", ""),
            "type": item.get("type", "transactional"),
            "category": item.get("category", "业务决策"),
            "participants": item.get("participants", []),
            "source_chat": item.get("source_chat", ""),
            "timestamp": item.get("timestamp", ""),
            "_chat_id": llm_output.get("chat_id", ""),
            "_chat_label": llm_output.get("chat_label", ""),
        }
        for item in items
        if isinstance(item, dict) and item.get("content")
    ]


def load_ndjson(path):
    chats = defaultdict(list)
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            m = json.loads(line)
            chats[m["chat_id"]].append(m)
    for cid in chats:
        chats[cid].sort(key=lambda x: x.get("create_time", ""))
    total = sum(len(v) for v in chats.values())
    print(f"  Loaded {total} messages across {len(chats)} chats", file=sys.stderr)
    return chats


def chat_label(messages):
    m0 = messages[0]
    label = m0.get("chat_name") or ""
    if not label and m0.get("chat_type") == "p2p":
        p = m0.get("chat_partner", {})
        label = p.get("name") or p.get("open_id", "")[:12]
    return label


def load_done_ids(raw_output):
    if not raw_output.exists():
        return set()
    done = set()
    with open(raw_output) as f:
        for line in f:
            line = line.strip()
            if line:
                done.add(json.loads(line)["chat_id"])
    return done


async def extract_all(llm, system_prompt, chats, raw_output):
    done_ids = load_done_ids(raw_output)
    total = len(chats)
    done_count = len(done_ids)
    print(f"  {done_count}/{total} chats already analyzed", file=sys.stderr)
    pending = [(cid, msgs) for cid, msgs in chats.items() if cid not in done_ids]
    if not pending:
        print("  All done, skipping", file=sys.stderr)
        return
    sem = asyncio.Semaphore(CONCURRENCY)

    async def process_one(chat_id, msgs):
        async with sem:
            label = chat_label(msgs)
            result = await analyze_one(llm, system_prompt, chat_id, label, msgs)
            with open(raw_output, "a") as f:
                f.write(json.dumps(result, ensure_ascii=False) + "\n")
            return result

    tasks = [process_one(cid, msgs) for cid, msgs in pending]
    for coro in asyncio.as_completed(tasks):
        r = await coro
        done_count += 1
        n = len(parse_consensus_from_output(r))
        print(f"  [{done_count}/{total}] {r['chat_label']:20s} {n} consensus(s)", file=sys.stderr)


def clean_participants(participants):
    return [p for p in participants if not p.startswith("ou_") and not p.startswith("oc_")]


def normalize(text):
    return re.sub(r"\s+", "", text.strip().lower())[:80]


def merge(results):
    all_items = []
    for r in results:
        for item in parse_consensus_from_output(r):
            item["participants"] = clean_participants(item.get("participants", []))
            all_items.append(item)

    merged = []
    for item in all_items:
        found = False
        for existing in merged:
            if normalize(existing["content"]) == normalize(item["content"]):
                found = True
                existing_p = set(existing.get("participants", []))
                existing_p.update(item.get("participants", []))
                existing["participants"] = sorted(existing_p)
                src = item.get("source_chat", "")
                if src and src not in existing.get("_srcs", [existing.get("source_chat", "")]):
                    existing.setdefault("_srcs", [existing.get("source_chat", "")])
                    if src not in existing["_srcs"]:
                        existing["_srcs"].append(src)
                    existing["source_chat"] = "、".join(existing["_srcs"])
                break
        if not found:
            item["participants"] = sorted(set(item.get("participants", [])))
            merged.append(item)

    merged.sort(key=lambda x: x.get("timestamp", ""))
    return merged


def main():
    parser = argparse.ArgumentParser(description="Extract consensus from chat records")
    parser.add_argument("--month", default="2026-07", help="Month label (YYYY-MM)")
    parser.add_argument("--data-dir", default="./data", help="Directory with chat-records-{month}.ndjson")
    parser.add_argument("--founder-name", default="Founder", help="Founder name in chat records")
    parser.add_argument("--input-prefix", default="chat-records", help="Input file prefix")
    parser.add_argument("--output-prefix", default="consensus", help="Output file prefix")
    args = parser.parse_args()

    data_dir = Path(args.data_dir)
    ndjson_file = data_dir / f"{args.input_prefix}-{args.month}.ndjson"
    raw_output = data_dir / f"{args.output_prefix}-raw-{args.month}.ndjson"
    json_output = data_dir / f"{args.output_prefix}-{args.month}.json"

    if not ndjson_file.exists():
        print(f"Input not found: {ndjson_file}", file=sys.stderr)
        sys.exit(1)

    print(f"Consensus extraction — {args.month}", file=sys.stderr)
    print(f"  Input:  {ndjson_file}", file=sys.stderr)
    print(f"  Output: {raw_output}\n          {json_output}", file=sys.stderr)
    print(file=sys.stderr)

    system_prompt = build_system_prompt(args.founder_name)
    chats = load_ndjson(ndjson_file)

    llm = AsyncLLM(model="deepseek-v4-flash")
    asyncio.run(extract_all(llm, system_prompt, chats, raw_output))

    results = []
    with open(raw_output) as f:
        for line in f:
            line = line.strip()
            if line:
                results.append(json.loads(line))

    all_consensuses = merge(results)
    output = {
        "month": args.month,
        "founder": args.founder_name,
        "generated_at": datetime.now().isoformat(),
        "total_consensuses": len(all_consensuses),
        "consensuses": all_consensuses,
    }
    with open(json_output, "w") as f:
        json.dump(output, f, ensure_ascii=False, indent=2)

    print(f"\nTotal: {len(all_consensuses)} consensus items", file=sys.stderr)
    print(f"Saved: {json_output}", file=sys.stderr)


if __name__ == "__main__":
    main()
