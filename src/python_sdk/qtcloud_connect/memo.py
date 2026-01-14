"""
Memo 领域模型
"""

from dataclasses import dataclass
from enum import Enum
from typing import Optional
from datetime import datetime


class MemoStatus(Enum):
    """
    备忘状态
    """
    DRAFT = "draft"
    CONFIRMED = "confirmed"
    ARCHIVED = "archived"


@dataclass
class Memo:
    """备忘"""
    id: str
    content: str
    status: Optional[MemoStatus] = None
    timestamp: Optional[datetime] = None
