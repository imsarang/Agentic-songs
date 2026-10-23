from __future__ import annotations

import asyncio
import json

from fastapi import FastAPI
from fastapi.responses import StreamingResponse

from shanks_agent.api.schemas import MCPCallRequest, TurnRequest
from shanks_agent.mcp.session import MCPSession, format_mcp_error

app = FastAPI(title="shanks-agent", version="0.1.0")


@app.get("/health")
async def health() -> dict:
    return {"ok": True, "service": "shanks-agent"}


def _sse(data: dict) -> str:
    return f"data: {json.dumps(data)}\n\n"


@app.post("/turn")
async def turn(body: TurnRequest) -> StreamingResponse:
    async def events():
        yield _sse(
            {
                "type": "tool_start",
                "name": "stub_echo",
                "args": {"text": body.text},
                "thread_id": body.thread_id,
            }
        )
        await asyncio.sleep(0.3)
        yield _sse(
            {
                "type": "tool_end",
                "name": "stub_echo",
                "ok": True,
                "preview": body.text[:200],
            }
        )
        yield _sse({"type": "assistant_message", "text": f"(stub) got: {body.text}"})
        yield _sse({"type": "turn_done"})

    return StreamingResponse(
        events(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


@app.get("/mcp/tools")
async def mcp_tools() -> dict:
    try:
        names = await MCPSession().list_tools()
        return {"ok": True, "tools": names}
    except Exception as exc:  # noqa: BLE001 — surface to HTTP client
        return {"ok": False, "error": format_mcp_error(exc)}


@app.post("/mcp/call")
async def mcp_call(body: MCPCallRequest) -> dict:
    try:
        preview = await MCPSession().call_tool(body.name, body.arguments)
        return {"ok": True, "preview": preview}
    except Exception as exc:  # noqa: BLE001
        return {"ok": False, "error": format_mcp_error(exc)}
