"""
qtcloud-connect provider — FastAPI 后端

基于 quanttide-connect 基础库，提供 REST API 供 Flutter 前端调用。
"""

from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI

from app.api import chat, consensuses, messages


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    # 启动时初始化
    yield
    # 关闭时清理
    pass


app = FastAPI(
    title="qtcloud-connect",
    description="量潮沟通云 — 人机沟通共识引擎 API",
    version="0.1.0",
    lifespan=lifespan,
)

app.include_router(chat.router, prefix="/api/v1", tags=["chat"])
app.include_router(messages.router, prefix="/api/v1", tags=["messages"])
app.include_router(consensuses.router, prefix="/api/v1", tags=["consensuses"])


@app.get("/health")
async def health():
    return {"status": "ok"}
