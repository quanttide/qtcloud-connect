from dataclasses import dataclass
from typing import List, Optional

@dataclass
class Topic:
    """话题"""
    id: str
    name: str  # 如："价值观讨论"、"业务场景探索"
    description: Optional[str] = None
    keywords: List[str] = None  # 话题关键词
