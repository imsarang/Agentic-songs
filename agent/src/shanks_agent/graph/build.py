from __future__ import annotations

import logging
from typing import Any

from langgraph.graph import END, START, StateGraph

from shanks_agent.graph.nodes.agent import make_agent_node, seed_messages
from shanks_agent.graph.nodes.tools import TOOL_NAMES, build_tool_nodes
from shanks_agent.graph.state import AgentState
from shanks_agent.llm.failover import FailoverLLM, build_failover_llm
from shanks_agent.mcp.session import MCPSession

logger = logging.getLogger(__name__)

_GRAPH = None


def _route(state: AgentState) -> str:
    name = state.get("next_tool")
    if name and name in TOOL_NAMES:
        nxt = f"tool_{name}"
        logger.info("route next=%s", nxt)
        return nxt
    logger.info("route next=END")
    return END


def get_graph():
    global _GRAPH
    if _GRAPH is None:
        logger.info("compiling graph (once)")
        _GRAPH = build_graph()
    return _GRAPH


def build_graph(
    llm: FailoverLLM | None = None,
    mcp: MCPSession | None = None,
):
    llm = llm or build_failover_llm()
    mcp = mcp or MCPSession()
    tool_nodes = build_tool_nodes(mcp)

    g: StateGraph = StateGraph(AgentState)
    g.add_node("agent", make_agent_node(llm))
    for node_id, fn in tool_nodes.items():
        g.add_node(node_id, fn)

    g.add_edge(START, "agent")
    g.add_conditional_edges(
        "agent",
        _route,
        {**{nid: nid for nid in tool_nodes}, END: END},
    )
    for node_id in tool_nodes:
        g.add_edge(node_id, "agent")

    return g.compile()


async def run_turn(
    user_text: str,
    *,
    thread_id: str = "default",
    graph=None,
) -> list[dict[str, Any]]:
    graph = graph or get_graph()
    logger.info("turn start thread_id=%s text=%s", thread_id, user_text[:200])

    result = await graph.ainvoke(
        {
            "messages": seed_messages(user_text),
            "next_tool": None,
            "tool_args": {},
            "event_sink": [],
        },
        config={
            "configurable": {"thread_id": thread_id},
            "tags": ["shanks", "agent"],
            "metadata": {"thread_id": thread_id},
        },
    )

    events = list(result.get("event_sink") or [])
    if not any(e.get("type") == "assistant_message" for e in events):
        for msg in reversed(result.get("messages") or []):
            if getattr(msg, "type", None) == "ai":
                events.append({"type": "assistant_message", "text": str(msg.content)})
                break
    events.append({"type": "turn_done"})
    logger.info("turn done thread_id=%s events=%d", thread_id, len(events))
    return events
