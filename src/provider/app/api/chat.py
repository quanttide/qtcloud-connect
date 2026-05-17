"""
聊天 API：发送消息 → 消息智能体回复 → 共识智能体观察。
"""

from __future__ import annotations

from fastapi import APIRouter, Depends
from pydantic import BaseModel

from app.deps import Services, get_services

router = APIRouter()


class ChatRequest(BaseModel):
    message: str
    conversation_id: str | None = None


class ChatResponse(BaseModel):
    reply: str
    message_id: str
    conversation_id: str


@router.post("/chat", response_model=ChatResponse)
async def chat(
    request: ChatRequest, services: Services = Depends(get_services)
) -> ChatResponse:
    # 1. 存用户消息
    user_msg = services.msg_svc.send(
        request.message, __import__("quanttide_connect").models.Role.user
    )
    services.history.append({"role": "user", "content": user_msg.content})

    # 2. 消息智能体回复
    confirmed = [
        {"content": c.content, "id": c.id}
        for c in services.storage.list_consensuses()
        if c.status.value == "confirmed"
    ]
    reply = services.msg_agent.reply(user_msg.content, services.history[:-1], confirmed)
    agent_msg = services.msg_svc.send(
        reply, __import__("quanttide_connect").models.Role.agent
    )
    services.history.append({"role": "assistant", "content": reply})

    # 3. 共识智能体观察（后台执行，不阻塞响应）
    services.con_agent.observe(user_msg, agent_msg, services.history)

    return ChatResponse(
        reply=reply,
        message_id=user_msg.id,
        conversation_id=request.conversation_id or "default",
    )
