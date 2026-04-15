# 话题

## 定义

话题（Topic）是从对话中提取的核心主题，由 AI 辅助生成、人工协同维护。它是对话内容的认知抽象，用于组织、检索和追踪沟通脉络。

## 与 Context 的区别

| 维度 | Context | Topic |
|------|---------|-------|
| 生成方式 | AI 纯自动提取 | AI 提取 + 人工协同 |
| 维护主体 | AI | 人 |
| 内容稳定性 | 不可修改 | 可编辑、重组 |
| 命名规则 | 自动生成摘要 | 人工命名 |
| 用途 | 认知依据 | 组织、检索、导航 |

## 数据模型

```python
@dataclass
class Topic:
    """话题"""
    id: str
    name: str                    # 人工命名
    description: Optional[str]    # 人工描述
    context_id: Optional[str]     # 关联的 AI 提取语境
    status: TopicStatus           # 人工管理状态
    keywords: List[str]           # 人工标注关键词
    created_by: str               # 创建人
    timestamp: Optional[datetime]
```

## 使用场景

1. **话题命名**：基于 AI 提取的 Context，人工为对话内容命名（如"利他与利己的讨论"）

2. **话题归类**：将多个 Context 归类到同一个 Topic 下，形成主题聚类

3. **话题检索**：通过关键词、名称快速定位相关对话

4. **话题追踪**：了解某个话题在不同时间段的演进

## 工作流程

```
对话文本 → AI 提取 Context → 人工创建 Topic → 持续维护
                    ↓                    ↓
              认知依据              可编辑/重组
```

## 状态管理

- `ACTIVE`：活跃话题
- `ARCHIVED`：已归档
- `MERGED`：已合并到其他话题
