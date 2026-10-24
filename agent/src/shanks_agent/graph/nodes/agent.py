from __future__ import annotations

import json
import logging
import re
from typing import Any

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage

from shanks_agent.graph.nodes.tools import TOOL_NAMES
from shanks_agent.graph.state import AgentState
from shanks_agent.llm.failover import FailoverLLM

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = """You control Spotify via MCP tools. Call at most one tool per turn.
Tools:
{tools}
Reply with ONLY JSON (no markdown):
{{"action":"finish","text":"..."}}
or
{{"action":"tool","name":"<tool_name>","args":{{...}}}}

Play flow (strict):
1) spotify_search with the user's song text (limit 5).
2) Pick the row whose title best matches what the user asked.
3) spotify_play using search handles ONLY:
   {{"search_id":N,"result_index":I,"title":"exact title from that row"}}
   Do NOT call spotify_play with only "query" if you already have search results.
4) spotify_now_playing. If the title/artist is clearly the wrong song, search again
   or play a different result_index. Do not finish until it matches or you tried 2 alternatives.
5) Only then action=finish and tell the user what is playing.

Control: spotify_control action=pause|resume|next|previous.
Playlists: spotify_playlist action=list|create|add|play|update.
"""


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
    tools_blob = "\n".join(f"- {n}" for n in TOOL_NAMES)

    async def agent_node(state: AgentState) -> dict[str, Any]:
        n_msgs = len(state.get("messages") or [])
        logger.info("node=agent enter messages=%d", n_msgs)
        messages = [
            SystemMessage(content=SYSTEM_PROMPT.format(tools=tools_blob)),
            *list(state.get("messages") or []),
        ]
        ai = await llm.acomplete(messages)
        content = ai.content if isinstance(ai.content, str) else str(ai.content)
        decision = _parse(content)
        sink = list(state.get("event_sink") or [])
        logger.info("node=agent decision=%s", decision)

        if decision.get("action") == "tool":
            name = str(decision.get("name") or "")
            if name not in TOOL_NAMES:
                msg = f"Unknown tool {name!r}. Use: {', '.join(TOOL_NAMES)}"
                logger.warning("node=agent unknown_tool=%s → finish", name)
                sink.append({"type": "assistant_message", "text": msg})
                return {
                    "messages": [AIMessage(content=msg)],
                    "next_tool": None,
                    "event_sink": sink,
                }
            args = decision.get("args") if isinstance(decision.get("args"), dict) else {}
            logger.info("node=agent route_tool=%s args=%s", name, args)
            return {
                "messages": [AIMessage(content=content)],
                "next_tool": name,
                "tool_args": args,
                "event_sink": sink,
            }

        text = str(decision.get("text") or content)
        logger.info("node=agent finish text=%s", text[:200])
        sink.append({"type": "assistant_message", "text": text})
        return {
            "messages": [AIMessage(content=text)],
            "next_tool": None,
            "event_sink": sink,
        }

    return agent_node


def seed_messages(user_text: str) -> list:
    return [HumanMessage(content=user_text)]
