"""
备忘
"""

from dataclasses import dataclass
from enum import Enum


class MemoStatus(Enum):
    """
    备忘状态
    """
    DRAFT = "draft"
    CONFIRMED = "confirmed"
    ARCHIVED = "archived"


@dataclass
class Memo:
    status: MemoStatus
