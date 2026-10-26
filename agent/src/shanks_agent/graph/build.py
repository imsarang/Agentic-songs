from __future__ import annotations

import logging
from collections.abc import AsyncIterator
from typing import Any

from langgraph.graph import END, START, StateGraph
from langsmith import trace

from shanks_agent.graph.nodes.agent import make_agent_node, seed_messages
from shanks_agent.graph.nodes.tools import TOOL_NAMES, build_tool_nodes
from shanks_agent.graph.prompts import resolve_domain
from shanks_agent.graph.state import AgentState
from shanks_agent.llm.failover import FailoverLLM, build_failover_llm
from shanks_agent.observability import new_run_id

logger = logging.getLogger(__name__)

_GRAPH = None
_LLM: FailoverLLM | None = None


def _route(state: AgentState) -> str:
    name = state.get("next_tool")
    if name and name in TOOL_NAMES:
        nxt = f"tool_{name}"
        logger.info("route next=%s", nxt)
        return nxt
    logger.info("route next=END")
    return END


def get_graph():
    global _GRAPH, _LLM
    if _GRAPH is None:
        logger.info("compiling graph (once)")
        _LLM = build_failover_llm()
        _GRAPH = build_graph(_LLM)
    return _GRAPH


def build_graph(llm: FailoverLLM | None = None):
    llm = llm or build_failover_llm()
    tool_nodes = build_tool_nodes()

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
    run_id: str | None = None,
    domain: str | None = None,
    graph=None,
) -> AsyncIterator[dict[str, Any]]:
    """Yield UI events as each graph node emits them.

    Nodes call ``emit()`` during execution. ``stream_mode="custom"`` delivers
    those payloads while the node is still running, instead of waiting for
    ``ainvoke`` to finish and dumping a collected list.
    """
    graph = graph or get_graph()
    llm = _LLM or build_failover_llm()
    run_uuid = new_run_id(run_id)
    run_key = str(run_uuid)

    async with trace(
        "shanks-turn",
        run_type="chain",
        run_id=run_uuid,
        inputs={"text": user_text, "thread_id": thread_id},
        tags=["shanks"],
        metadata={"thread_id": thread_id, "run_id": run_key},
    ) as ls_run:
        yield {"type": "run", "run_id": run_key, "thread_id": thread_id}
        resolved = await resolve_domain(llm, user_text, domain)
        # Charts group by run name. One shared name draws one line.
        ls_run.name = f"shanks-{resolved}"
        ls_run.add_metadata({"domain": resolved})
        ls_run.add_tags([resolved])
        logger.info(
            "turn start run_id=%s thread_id=%s domain=%s text=%s",
            run_key,
            thread_id,
            resolved,
            user_text[:200],
        )
        yield {"type": "domain", "domain": resolved}

        saw_assistant = False
        last_values: dict[str, Any] | None = None
        n = 0
        async for mode, data in graph.astream(
            {
                "messages": seed_messages(user_text),
                "domain": resolved,
                "next_tool": None,
                "tool_args": {},
            },
            config={
                "configurable": {"thread_id": thread_id},
                "tags": ["shanks", "agent", resolved],
                "metadata": {
                    "thread_id": thread_id,
                    "run_id": run_key,
                    "domain": resolved,
                },
            },
            stream_mode=["custom", "values"],
        ):
            if mode == "values" and isinstance(data, dict):
                last_values = data
                continue
            if mode != "custom" or not isinstance(data, dict):
                continue
            if data.get("type") == "assistant_message":
                saw_assistant = True
            n += 1
            logger.info("stream event run_id=%s type=%s", run_key, data.get("type"))
            yield data

        if not saw_assistant and last_values is not None:
            for msg in reversed(last_values.get("messages") or []):
                if getattr(msg, "type", None) == "ai":
                    n += 1
                    yield {"type": "assistant_message", "text": str(msg.content)}
                    break
        n += 1
        yield {"type": "turn_done", "run_id": run_key}
        ls_run.end(outputs={"events": n, "domain": resolved})
        logger.info(
            "turn done run_id=%s thread_id=%s events=%d",
            run_key,
            thread_id,
            n,
        )
