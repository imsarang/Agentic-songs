from __future__ import annotations

import json

from fastapi import FastAPI
from fastapi.responses import StreamingResponse
from langchain_core.messages import HumanMessage

from shanks_agent.api.schemas import MCPCallRequest, TurnRequest
from shanks_agent.graph.build import get_graph, run_turn
from shanks_agent.llm import build_failover_llm
from shanks_agent.mcp.session import MCPSession, format_mcp_error

app = FastAPI(title="shanks-agent", version="0.1.0")

@app.on_event('startup')
async def _warmup_graph() -> None:
    get_graph()
    
@app.get("/health")
async def health() -> dict:
    return {"ok": True, "service": "shanks-agent"}


def _sse(data: dict) -> str:
    return f"data: {json.dumps(data)}\n\n"


@app.post("/turn")
async def turn(body: TurnRequest) -> StreamingResponse:
    async def events():
        try:
            async for event in run_turn(
                body.text,
                thread_id=body.thread_id,
                domain=body.domain,
            ):
                yield _sse(event)
        except Exception as exc:  # noqa: BLE001
            yield _sse({"type": "error", "message": format_mcp_error(exc)})
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
    except Exception as exc:  # noqa: BLE001
        return {"ok": False, "error": format_mcp_error(exc)}


@app.post("/mcp/call")
async def mcp_call(body: MCPCallRequest) -> dict:
    try:
        preview = await MCPSession().call_tool(body.name, body.arguments)
        return {"ok": True, "preview": preview}
    except Exception as exc:  # noqa: BLE001
        return {"ok": False, "error": format_mcp_error(exc)}


@app.post("/llm/ping")
async def llm_ping(text: str = "say hi in 3 words") -> dict:
    try:
        ai = await build_failover_llm().acomplete([HumanMessage(content=text)])
        return {"ok": True, "text": str(ai.content)}
    except Exception as exc:  # noqa: BLE001
        return {"ok": False, "error": str(exc)}
