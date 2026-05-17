"""
消息 API：查询消息历史。
"""

from __future__ import annotations

from fastapi import APIRouter, Depends

from app.deps import Services, get_services

router = APIRouter()


@router.get("/messages")
async def list_messages(services: Services = Depends(get_services)):
    messages = services.storage.list_messages()
    return [
        {
            "id": m.id,
            "content": m.content,
            "role": m.role.value,
            "created_at": m.created_at.isoformat(),
            "updated_at": m.updated_at.isoformat() if m.updated_at else None,
        }
        for m in messages
    ]


@router.get("/messages/{message_id}")
async def get_message(message_id: str, services: Services = Depends(get_services)):
    m = services.storage.get_message(message_id)
    if not m:
        return {"error": "not found"}, 404
    return {
        "id": m.id,
        "content": m.content,
        "role": m.role.value,
        "created_at": m.created_at.isoformat(),
        "updated_at": m.updated_at.isoformat() if m.updated_at else None,
    }
