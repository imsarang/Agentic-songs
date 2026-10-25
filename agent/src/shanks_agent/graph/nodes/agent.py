from __future__ import annotations

import json
import logging
import re
from typing import Any

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage

from shanks_agent.graph.emit import emit
from shanks_agent.graph.nodes.tools import DOMAIN_TOOLS
from shanks_agent.graph.prompts import build_system_prompt
from shanks_agent.graph.state import AgentState
from shanks_agent.llm.failover import FailoverLLM

logger = logging.getLogger(__name__)


def _parse(raw: str) -> dict[str, Any]:
    text = raw.strip()
    m = re.search(r"```(?:json)?\s*([\s\S]*?)```", text)
    if m:
        text = m.group(1).strip()
    try:
        data = json.loads(text)
        if isinstance(data, dict):
            return data
    except json.JSONDecodeError:
        pass
    return {"action": "finish", "text": raw}


def make_agent_node(llm: FailoverLLM):
    async def agent_node(state: AgentState) -> dict[str, Any]:
        domain = state.get("domain") or "spotify"
        allowed = DOMAIN_TOOLS.get(domain, DOMAIN_TOOLS["spotify"])
        n_msgs = len(state.get("messages") or [])
        logger.info("node=agent enter domain=%s messages=%d", domain, n_msgs)
        messages = [
            SystemMessage(content=build_system_prompt(domain, allowed)),
            *list(state.get("messages") or []),
        ]
        ai = await llm.acomplete(messages)
        # Anthropic may return [{thinking…}, {type:text, text:…}] — keep text only
        if isinstance(ai.content, list):
            content = "".join(
                b.get("text", "") for b in ai.content if isinstance(b, dict) and b.get("type") == "text"
            )
        else:
            content = str(ai.content or "")
        decision = _parse(content)
        logger.info("node=agent decision=%s", decision)

        if decision.get("action") == "tool":
            name = str(decision.get("name") or "")
            if name not in allowed:
                msg = f"Unknown tool {name!r} for {domain}. Use: {', '.join(allowed)}"
                logger.warning("node=agent unknown_tool=%s → finish", name)
                emit({"type": "assistant_message", "text": msg})
                return {
                    "messages": [AIMessage(content=msg)],
                    "next_tool": None,
                }
            args = decision.get("args") if isinstance(decision.get("args"), dict) else {}
            logger.info("node=agent route_tool=%s args=%s", name, args)
            return {
                "messages": [AIMessage(content=content)],
                "next_tool": name,
                "tool_args": args,
            }

        text = str(decision.get("text") or content)
        logger.info("node=agent finish text=%s", text[:200])
        emit({"type": "assistant_message", "text": text})
        return {
            "messages": [AIMessage(content=text)],
            "next_tool": None,
        }

    return agent_node


def seed_messages(user_text: str) -> list:
    return [HumanMessage(content=user_text)]
