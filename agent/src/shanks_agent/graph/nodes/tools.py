from __future__ import annotations

import logging
from collections.abc import Awaitable, Callable
from typing import Any

from langchain_core.messages import HumanMessage

from shanks_agent.config import get_settings
from shanks_agent.graph.emit import emit
from shanks_agent.graph.state import AgentState
from shanks_agent.mcp.session import MCPSession

logger = logging.getLogger(__name__)

SPOTIFY_TOOLS: tuple[str, ...] = (
    "spotify_search",
    "spotify_play",
    "spotify_control",
    "spotify_now_playing",
    "spotify_playlist",
)

MACHINE_TOOLS: tuple[str, ...] = (
    "machine_spec",
    "machine_capacity",
    "machine_processes",
    "machine_cgroups",
    "machine_namespaces",
    "machine_net",
    "machine_blockers",
    "machine_limits",
    "machine_power",
    "machine_gpu",
    "machine_logs",
    "machine_health",
    "machine_grep",
    "machine_cat",
)

GOOGLE_TOOLS: tuple[str, ...] = (
    "google_switch_account",
    "google_create_subject",
    "google_get_receiver",
    "google_create_draft",
    "google_send_mail",
    "google_search_inbox",
    "google_search_social",
    "google_search_promotions",
    "google_search_spam",
    "google_search_sent",
    "google_search_drafts",
    "google_delete_drafts",
    "google_get_message",
    "google_delete_mail",
)

CREATIVE_TOOLS: tuple[str, ...] = ()


DRIVE_TOOLS = ("google_drive_space", "google_drive")

DOMAIN_TOOLS: dict[str, tuple[str, ...]] = {
    "spotify": SPOTIFY_TOOLS,
    "machine": MACHINE_TOOLS,
    "google": GOOGLE_TOOLS,
    "creative": CREATIVE_TOOLS,
    "drive": DRIVE_TOOLS
}

TOOL_NAMES: tuple[str, ...] = SPOTIFY_TOOLS + MACHINE_TOOLS + GOOGLE_TOOLS

ToolNode = Callable[[AgentState], Awaitable[dict[str, Any]]]


def make_tool_node(tool_name: str, mcp: MCPSession) -> ToolNode:
    async def _node(state: AgentState) -> dict[str, Any]:
        args = dict(state.get("tool_args") or {})
        who = str(state.get("domain") or "")
        emit({"type": "tool_start", "subagent": who, "name": tool_name, "args": args})
        logger.info("node=tool_%s start args=%s", tool_name, args)
        try:
            preview = await mcp.call_tool(tool_name, args)
            ok = True
            logger.info("node=tool_%s ok preview=%s", tool_name, preview[:300])
        except Exception as exc:  # noqa: BLE001
            preview = str(exc)
            ok = False
            logger.exception("node=tool_%s failed: %s", tool_name, exc)
        emit({"type": "tool_end", "subagent": who, "name": tool_name, "ok": ok})
        head, sep, body = preview.partition("\n---\n")
        if sep:
            emit({"type": "assistant_message", "subagent": who, "text": body})
            preview = head
        return {
            "messages": [HumanMessage(content=f"[tool {tool_name} result]\n{preview}")],
            "next_tool": None,
            "tool_args": {},
        }

    _node.__name__ = f"tool_{tool_name}"
    return _node


def build_tool_nodes(
    spotify_mcp: MCPSession | None = None,
    machine_mcp: MCPSession | None = None,
    google_mcp: MCPSession | None = None,
) -> dict[str, ToolNode]:
    settings = get_settings()
    spotify_mcp = spotify_mcp or MCPSession(settings.mcp_url)
    machine_mcp = machine_mcp or MCPSession(settings.machine_mcp_url)
    google_mcp = google_mcp or MCPSession(settings.google_mcp_url)
    nodes: dict[str, ToolNode] = {}
    for name in SPOTIFY_TOOLS:
        nodes[f"tool_{name}"] = make_tool_node(name, spotify_mcp)
    for name in MACHINE_TOOLS:
        nodes[f"tool_{name}"] = make_tool_node(name, machine_mcp)
    for name in GOOGLE_TOOLS:
        nodes[f"tool_{name}"] = make_tool_node(name, google_mcp)
    return nodes
