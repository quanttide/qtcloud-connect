#!/usr/bin/env python3
"""
Example: Analyze chat records to find team-driven business activities
and trace their flow from data input to business feedback.

Pipeline:
  Step 1: Load NDJSON -> keyword filter -> save business segments
  Step 2: Parallel LLM analysis on each segment (incremental save)
  Step 3: Merge, deduplicate, classify -> final JSON + Markdown report

Usage:
    python3 analyze-team-activities.py --ndjson chat-records-2026-07.ndjson \
        --founder-name "张果" --month 2026-07

Requires:
    pip install quanttide-agent
    LLM API key configured (env: QUANTTIDE_LLM_API_KEY)
"""

import asyncio, json, os, re, sys, argparse
from collections import defaultdict
from datetime import datetime

from quanttide_agent.llm import AsyncLLM
from quanttide_agent.message import Message

DEFAULT_BUSINESS_KEYWORDS = [
    "合同", "客户", "报价", "方案", "项目", "合作", "签约", "订单",
    "预算", "付款", "收款", "发票", "采购", "供应商", "商务",
    "产品", "上线", "发布", "迭代", "需求", "排期", "交付",
    "招聘", "入职", "面试", "offer",
    "数据", "分析", "报告", "指标", "增长", "营收", "利润",
    "会议", "决策", "审批", "流程", "制度", "规范",
    "市场", "营销", "推广", "渠道", "用户", "转化",
    "融资", "投资", "股权", "估值",
    "战略", "规划", "目标", "OKR", "KPI",
]

CONCURRENCY = 5


def build_system_prompt(founder_name):
    return f"""你是一个业务分析师，正在分析公司内部聊天记录。

## 背景
- 数据源是创始人「{founder_name}」的聊天记录，你只能看到{founder_name}参与或可见的对话
- 创始人是「{founder_name}」（聊天记录中 sender.name == "{founder_name}"）
- 「团队」= 除{founder_name}以外的成员
- 「业务活动」= 具有商业目的的行动：客户对接、方案讨论、产品决策、对外协作、资源协调等
- 排除：纯粹闲聊、纯技术细节、仅信息同步无后续行动

## 任务
对于这段聊天记录中的每条业务活动，提取：

1. **activity**: 活动名称
2. **driver**: 从{founder_name}视角看，发起或主要推动该活动的人
3. **trigger**: 触发/信息输入——是什么启动了这件事？
4. **actions**: 行动过程——相关方做了什么？
5. **feedback**: 商业反馈/结果——产生了什么商业结果？
6. **flow**: 一句话描述完整链路「信息输入 → 行动 → 商业反馈」

## 输出格式
只返回 JSON 数组，不要其他文字：
[{{"activity": "...", "driver": "...", "trigger": "...", "actions": "...", "feedback": "...", "flow": "..."}}]

如果没有，返回 []"""


def load_messages(path):
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
    return dict(chats)


def chat_label(messages):
    m = messages[0]
    name = m.get("chat_name") or ""
    if not name and m.get("chat_type") == "p2p":
        p = m.get("chat_partner", {})
        name = p.get("name") or p.get("open_id", "")[:12]
    return name or m["chat_id"][:12]


def filter_business_segments(chats, keywords):
    all_segments = []
    for cid, msgs in chats.items():
        label = chat_label(msgs)
        current = []
        for m in msgs:
            content = (m.get("content") or "").lower()
            has_kw = any(kw in content for kw in keywords)
            if has_kw:
                current.append(m)
            else:
                if len(current) >= 3:
                    all_segments.append({
                        "chat_id": cid, "chat_label": label,
                        "messages": current, "msg_count": len(current),
                        "time_start": current[0].get("create_time", ""),
                        "time_end": current[-1].get("create_time", ""),
                    })
                current = []
        if len(current) >= 3:
            all_segments.append({
                "chat_id": cid, "chat_label": label,
                "messages": current, "msg_count": len(current),
                "time_start": current[0].get("create_time", ""),
                "time_end": current[-1].get("create_time", ""),
            })
    return all_segments


def render_messages(messages):
    lines = []
    for m in messages:
        ts = m.get("create_time", "")
        sender = m.get("sender", {})
        name = sender.get("name", sender.get("id", "?")[:12])
        content = (m.get("content") or "")[:600]
        lines.append(f"[{ts}] {name}: {content}")
    return "\n".join(lines)


async def analyze_one(llm, system_prompt, seg):
    chat_text = render_messages(seg["messages"])
    prompt = f"会话：{seg['chat_label']}\n时间：{seg['time_start']} ~ {seg['time_end']}\n\n{chat_text}"
    resp = await llm.complete(
        [Message(role="system", content=system_prompt), Message(role="user", content=prompt)],
        model="deepseek-v4-flash", temperature=0.1, max_tokens=2000,
    )
    return {"chat_id": seg["chat_id"], "chat_label": seg["chat_label"],
            "time_start": seg["time_start"], "time_end": seg["time_end"],
            "llm_output": resp.content}


async def step2_analyze(segments, results_file, system_prompt):
    done_ids = set()
    if os.path.exists(results_file):
        with open(results_file) as f:
            for line in f:
                line = line.strip()
                if line:
                    done_ids.add(json.loads(line)["_segment_idx"])
    pending = [(i, seg) for i, seg in enumerate(segments) if i not in done_ids]
    if not pending:
        print("  All segments already analyzed.", file=sys.stderr)
        return

    llm = AsyncLLM(model="deepseek-v4-flash")
    sem = asyncio.Semaphore(CONCURRENCY)

    async def bounded(idx, seg):
        async with sem:
            try:
                result = await analyze_one(llm, system_prompt, seg)
                result["_segment_idx"] = idx
                with open(results_file, "a") as f:
                    f.write(json.dumps(result, ensure_ascii=False) + "\n")
                return result
            except Exception as e:
                print(f"  [ERROR] segment {idx}: {e}", file=sys.stderr)
                return None

    tasks = [bounded(idx, seg) for idx, seg in pending]
    done_count = len(done_ids)
    for coro in asyncio.as_completed(tasks):
        r = await coro
        if r:
            done_count += 1
            print(f"  [{done_count}/{len(segments)}] {r['chat_label']}: {len(r.get('llm_output',''))} chars", file=sys.stderr)
    print(f"  Done. Results saved to {results_file}", file=sys.stderr)


def step3_merge(segments_file, results_file, output_json, output_md, founder_name, month):
    results = []
    if os.path.exists(results_file):
        with open(results_file) as f:
            for line in f:
                line = line.strip()
                if line:
                    results.append(json.loads(line))
    print(f"  Loaded {len(results)} segment analyses", file=sys.stderr)

    activities = []
    seen = set()
    for r in results:
        raw = r.get("llm_output", "")
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
            continue
        for item in items:
            if not isinstance(item, dict) or not item.get("activity"):
                continue
            key = item["activity"].strip().lower()[:50]
            if key and key not in seen:
                seen.add(key)
                item["_chat"] = r.get("chat_label", "?")
                item["_chat_id"] = r.get("chat_id", "?")
                item["_type"] = "founder_led" if item.get("driver", "") == founder_name else "team_active"
                activities.append(item)

    activities.sort(key=lambda x: (x.get("_type", ""), x.get("_chat", "")))

    # Save JSON
    output = {"month": month, "note": f"Data source: {founder_name}'s chat records. Findings reflect only what {founder_name} can see.",
              "total_activities": len(activities), "activities": activities}
    with open(output_json, "w") as f:
        json.dump(output, f, ensure_ascii=False, indent=2)
    print(f"  Saved: {output_json} ({len(activities)} activities)", file=sys.stderr)

    # Generate Markdown report
    team_count = sum(1 for a in activities if a["_type"] == "team_active")
    founder_count = sum(1 for a in activities if a["_type"] == "founder_led")

    lines = [
        f"# Business Activity Analysis — {month}",
        "",
        f"> Data source: {founder_name}'s chat records. Findings reflect only observable conversations.",
        "",
        f"**{len(activities)} activities found** — 🤖 Team-active: {team_count}  |  👤 {founder_name}-led: {founder_count}",
        "", "---", "",
    ]

    def render(i, a, label):
        return [
            f"### {i}. {a['activity']}", "",
            f"**{label}** | Chat: {a['_chat']} | Driver (in view): {a['driver']}", "",
            "| Stage | Content |", "|-------|---------|",
            f"| 📥 Trigger/Input | {a.get('trigger','?')} |",
            f"| 🔧 Actions | {a.get('actions','?')} |",
            f"| 💰 Business Feedback | {a.get('feedback','?')} |", "",
            "**Flow:**", "", f"> {a.get('flow','?')}", "", "---", "",
        ]

    idx = 0
    for a in activities:
        if a["_type"] == "team_active":
            idx += 1
            lines.extend(render(idx, a, "🤖 Team-active"))
    for a in activities:
        if a["_type"] == "founder_led":
            idx += 1
            lines.extend(render(idx, a, f"👤 {founder_name}-led"))

    with open(output_md, "w") as f:
        f.write("\n".join(lines))
    print(f"  Saved: {output_md}", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(description="Analyze chat records for team-driven business activities")
    parser.add_argument("--ndjson", required=True, help="Input NDJSON file")
    parser.add_argument("--month", default="YYYY-MM", help="Month label (for output files)")
    parser.add_argument("--founder-name", default="Founder", help="Founder's name in chat records")
    parser.add_argument("--output-prefix", default="team-driven-activities", help="Output file prefix")
    parser.add_argument("--keywords", nargs="*", default=DEFAULT_BUSINESS_KEYWORDS, help="Business keywords for filtering")
    args = parser.parse_args()

    month = args.month
    seg_file = f"step1-segments-{month}.json"
    res_file = f"step2-results-{month}.ndjson"
    out_json = f"{args.output_prefix}-{month}.json"
    out_md = f"{args.output_prefix}-{month}.md"

    print(f"Pipeline: {args.ndjson} -> {seg_file} -> {res_file} -> {out_json} + {out_md}", file=sys.stderr)

    # Step 1
    chats = load_messages(args.ndjson)
    total = sum(len(v) for v in chats.values())
    segments = filter_business_segments(chats, args.keywords)
    seg_count = sum(s["msg_count"] for s in segments)
    print(f"Step 1: {total} msgs -> {len(segments)} segments ({seg_count} msgs)", file=sys.stderr)
    with open(seg_file, "w") as f:
        json.dump({"month": month, "total_messages": total, "segments_count": len(segments), "segments": segments}, f, ensure_ascii=False, indent=2)

    # Step 2
    system_prompt = build_system_prompt(args.founder_name)
    asyncio.run(step2_analyze(segments, res_file, system_prompt))

    # Step 3
    step3_merge(seg_file, res_file, out_json, out_md, args.founder_name, month)
    print("Done.", file=sys.stderr)


if __name__ == "__main__":
    main()
