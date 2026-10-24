from __future__ import annotations

import logging
from collections.abc import Awaitable, Callable
from typing import Any

from langchain_core.messages import HumanMessage

from shanks_agent.graph.state import AgentState
from shanks_agent.mcp.session import MCPSession

logger = logging.getLogger(__name__)

TOOL_NAMES: tuple[str, ...] = (
    "spotify_search",
    "spotify_play",
    "spotify_control",
    "spotify_now_playing",
    "spotify_playlist",
)

ToolNode = Callable[[AgentState], Awaitable[dict[str, Any]]]


def make_tool_node(tool_name: str, mcp: MCPSession) -> ToolNode:
    async def _node(state: AgentState) -> dict[str, Any]:
        args = dict(state.get("tool_args") or {})
        sink = list(state.get("event_sink") or [])
        sink.append({"type": "tool_start", "name": tool_name, "args": args})
        logger.info("node=tool_%s start args=%s", tool_name, args)
        try:
            preview = await mcp.call_tool(tool_name, args)
            ok = True
            logger.info("node=tool_%s ok preview=%s", tool_name, preview[:300])
        except Exception as exc:  # noqa: BLE001
            preview = str(exc)
            ok = False
            logger.exception("node=tool_%s failed: %s", tool_name, exc)
        sink.append(
            {"type": "tool_end", "name": tool_name, "ok": ok, "preview": preview[:800]}
        )
        return {
            "messages": [HumanMessage(content=f"[tool {tool_name} result]\n{preview}")],
            "next_tool": None,
            "tool_args": {},
            "event_sink": sink,
        }

    _node.__name__ = f"tool_{tool_name}"
    return _node


def build_tool_nodes(mcp: MCPSession) -> dict[str, ToolNode]:
    return {f"tool_{name}": make_tool_node(name, mcp) for name in TOOL_NAMES}
