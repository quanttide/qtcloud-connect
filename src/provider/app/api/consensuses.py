"""
共识 API：查询、确认、废弃共识。
"""

from __future__ import annotations

from fastapi import APIRouter, Depends
from pydantic import BaseModel

from app.deps import Services, get_services

router = APIRouter()


class ConfirmRequest(BaseModel):
    consensus_id: str


class DeprecateRequest(BaseModel):
    consensus_id: str


@router.get("/consensuses")
async def list_consensuses(services: Services = Depends(get_services)):
    consensuses = services.storage.list_consensuses()
    result = []
    for c in consensuses:
        rels = services.storage.get_relations_for_consensus(c.id)
        result.append(
            {
                "id": c.id,
                "content": c.content,
                "status": c.status.value,
                "created_at": c.created_at.isoformat(),
                "related_message_ids": [r.message_id for r in rels],
            }
        )
    return result


@router.post("/consensuses/confirm")
async def confirm_consensus(
    req: ConfirmRequest, services: Services = Depends(get_services)
):
    c = services.con_svc.confirm(req.consensus_id)
    if not c:
        return {"error": "not found"}, 404
    return {"id": c.id, "status": c.status.value}


@router.post("/consensuses/deprecate")
async def deprecate_consensus(
    req: DeprecateRequest, services: Services = Depends(get_services)
):
    c = services.con_svc.deprecate(req.consensus_id)
    if not c:
        return {"error": "not found"}, 404
    return {"id": c.id, "status": c.status.value}
