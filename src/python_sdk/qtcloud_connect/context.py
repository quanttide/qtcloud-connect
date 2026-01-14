"""
Context 领域模型
"""

from dataclasses import dataclass
from typing import Optional
from datetime import datetime


@dataclass
class Context:
    """认知语境"""
    id: str
    summary: str
    shared_knowledge: Optional[str] = None
    mental_schemas: Optional[str] = None
    psychological_assumptions: Optional[str] = None
    timestamp: Optional[datetime] = None
