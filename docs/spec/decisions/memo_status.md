# 备忘状态

## 已确认：`agreed` v.s. `confirmed`

“agreed” 听起来很合理——毕竟“已确认”常常意味着“大家同意了”，但在产品协作系统中，agreed 作为状态标签不如 confirmed 稳妥，原因如下：

🔍 核心区别：语义重心不同

词 语义重心 隐含问题
---- -------- --------
agreed 达成一致的过程（强调“共识”） → 谁和谁 agree？需要多少人 agree？口头 agree 算吗？
confirmed 结果已被核实/锁定（强调“确定性”） → 无论怎么达成的，现在它是“有效输入”

在无账号、无审批流的轻量系统中，“agreed” 容易引发模糊性，而 “confirmed” 更聚焦结果。

⚠️ 为什么 agreed 在实践中可能带来问题？
1. 主体缺失（尤其在无账号系统中）
“Agreed” 是一个及物动作：A agreed with B。
但你的系统不知道“谁”同意了——是作者自说自话？还是团队真共识？
结果：有人可能把“我觉得OK”标为 agreed，造成误判。
✅ Confirmed 则更中性：它不追问“谁确认”，只表示“当前内容被视为有效”。
2. 门槛模糊
“Agreed” 暗示多方共识，但现实中：
有些需求只需 PM + Tech Lead 对齐；
有些需跨部门评审。
如果状态叫 agreed，团队会纠结：“到底几个人同意才算？”
✅ Confirmed 只要求内容本身具备可执行性，不绑定参与人数。
3. 与工程/产品术语习惯不符
主流工具几乎不用 agreed 作状态：

工具 常见状态
------ --------
Jira Approved, Confirmed, Done
Linear Triage, Planned, Completed
Notion 模板 Draft, In Review, Confirmed, Archived
GitHub Issues Open, Closed, Accepted

→ Confirmed 是行业通用语，降低认知成本。
4. AI 或自动化处理更可靠
AI 可通过关键词判断是否“confirmed”：
“We will go with…”
“Final decision: …”
“Confirmed on Jan 5”
但“agreed”常出现在非结论句中：
“Some people agreed, but others didn’t.” ← 这显然不该标为 agreed

✅ 什么时候可以用 agreed？
如果你有明确的投票/审批机制（如：3/5 人点击“agree”才生效）；
或在法律/合同场景（“双方已签署 agreement”）；
或作为动词描述过程（“We agreed on the scope”）。

但在状态标签（status label） 场景下，confirmed 更安全、更通用、更结果导向。

📝 类比理解

中文说法 推荐英文状态 为什么不选另一个
-------- ------------- ----------------
“这个方案定了” ✅ Confirmed ❌ Agreed → 谁定的？
“会议达成一致” （描述用）Agreed （但状态仍标 Confirmed）
“需求已冻结” ✅ Confirmed / Finalized ❌ Agreed → 冻结是结果，不是过程

💡 建议
坚持用 confirmed 作为状态值。
如果想强调“团队共识”，可以在文档正文中写 “Team agreed to proceed with...”，但状态字段保持为 confirmed ——这样既保留语义，又保证系统简洁。
状态是给系统看的，正文是给人看的。

希望这能帮你避开未来的一个小坑！