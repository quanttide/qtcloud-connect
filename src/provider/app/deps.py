"""
共享依赖：Storage、Services、Agents。

每个请求创建独立实例（SQLite 连接为每个请求新建），
历史状态暂存于内存（后续可替换为 Redis）。
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import AsyncGenerator

from fastapi import Depends
from quanttide_connect.services.consensus import ConsensusService
from quanttide_connect.services.message import MessageService
from quanttide_connect.services.relation import RelationService

from app.agents.consensus_agent import ConsensusAgent
from app.agents.message_agent import MessageAgent
from app.config import settings
from app.storage import Storage


def get_storage() -> Storage:
    return Storage()


@dataclass
class Services:
    storage: Storage
    msg_svc: MessageService
    con_svc: ConsensusService
    rel_svc: RelationService
    msg_agent: MessageAgent
    con_agent: ConsensusAgent
    history: list[dict] = field(default_factory=list)


def get_services(
    storage: Storage = Depends(get_storage),
) -> AsyncGenerator[Services, None]:
    msg_svc = MessageService(storage)
    con_svc = ConsensusService(storage)
    rel_svc = RelationService(storage)
    msg_agent = MessageAgent()
    con_agent = ConsensusAgent(storage, con_svc, rel_svc)

    yield Services(
        storage=storage,
        msg_svc=msg_svc,
        con_svc=con_svc,
        rel_svc=rel_svc,
        msg_agent=msg_agent,
        con_agent=con_agent,
    )
