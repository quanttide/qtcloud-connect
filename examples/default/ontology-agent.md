# 本体


目标：构建一个轻量、松耦合、面向语境的记忆与交互基础模型

🔹 1. 命名空间统一
- 前缀：connect:
- URI：urn:qtcloud:connect:
- 所有概念基于此命名空间定义，确保跨文件一致性。

🔹 2. 核心设计原则确认
原则   说明
分层架构   主体 → 媒介 → 语境，三层解耦

松耦合   语境（Context）不依赖原始媒介（Medium）的具体存在

语境中心   Context 是系统唯一需要持久化的语义单元，承担“记忆”角色

极简本体   避免过度建模，仅保留驱动行为差异的实体

🔹 3. 最终保留的核心概念（3个）
实体   作用   关键属性
connect:Agent   沟通主体   AI, Human 子类

connect:Medium   原始交互媒介   sentBy, content, belongsToStream, mediumTimestamp（支持消息、语音、文档等子类型）

connect:Context   语义上下文（核心！）   hasParticipant, summary, timestamp, derivedFromStream

🎯 Context 同时作为“上下文”和“记忆”的载体，无需独立 MemoryRecord。

🔹 4. 明确移除的概念
概念   移除原因
Scenario（场景类）   尚未驱动系统行为差异，可用 contextType 字符串替代

MemoryRecord   功能由 Context 承担，避免冗余实体

强引用关系（如 recordedBy）   改为通过流 ID 或时间窗口间接关联，保持松耦合

🔹 5. 松耦合机制实现

- 媒介 ↔ 语境 通过 共享流标识符 关联：
  - Medium.belongsToStream = "stream-xyz"
  - Context.derivedFromStream = "stream-xyz"
- 不要求 Medium 实例存在于 RDF 图中，Context 可独立存在。
- 支持基于时间窗口的弱关联（未来扩展）。

🔹 6. 推荐文件结构

docs/knowl/ontology/connect/
├── agent.ttl      # 沟通主体（AI / Human）
├── medium.ttl     # 交互媒介（消息、语音、文件等）
└── context.ttl    # 语境（核心语义单元，即“记忆”）

每个文件可独立使用，也可合并加载到同一 RDF 图中自动关联。

🔹 7. 下一步建议方向

- 在应用层实现 Medium → Context 的自动提炼管道（如 LLM 摘要 + 参与者识别）
- 设计 Context 的 生命周期管理策略（合并、过期、版本控制）
- 将 Context.summary 向量化，接入向量数据库实现语义检索
- 通过 contextType 支持不同类型语境（如 "user-preference", "project-context"）

