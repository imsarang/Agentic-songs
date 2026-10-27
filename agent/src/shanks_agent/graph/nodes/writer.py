from __future__ import annotations

import json
import logging
import re

from langchain_core.messages import AIMessage, SystemMessage

from shanks_agent.graph.emit import emit
from shanks_agent.graph.prompts.creative import CREATIVE
from shanks_agent.graph.state import AgentState
from shanks_agent.llm.failover import FailoverLLM

logger = logging.getLogger(__name__)

_JSON = re.compile(r"\{[\s\S]*\}")


def make_writer_node(llm: FailoverLLM):
    async def write(state: AgentState) -> dict:
        messages = [SystemMessage(content=CREATIVE), *list(state.get("messages") or [])]
        ai = await llm.acomplete(messages)
        raw = ai.content
        if isinstance(raw, list):
            raw = "".join(
                b.get("text", "") for b in raw if isinstance(b, dict) and b.get("type") == "text"
            )
        raw = str(raw or "").strip()
        data = _load(raw)
        subject = str(data.get("subject") or "").strip()
        body = str(data.get("body") or "").strip()
        if subject or body:
            shown = "\n\n".join(part for part in (f"Subject: {subject}" if subject else "", body) if part)
            payload = json.dumps({"subject": subject, "body": body})
        else:
            shown = str(data.get("text") or raw).strip()
            payload = shown
        logger.info("node=creative subject=%s chars=%d", bool(subject), len(shown))
        emit({"type": "assistant_message", "subagent": "creative", "text": shown})
        return {"messages": [AIMessage(content=payload)], "next_tool": None}

    return write


def draft_from(values: dict | None) -> str:
    if not values:
        return ""
    for msg in reversed(values.get("messages") or []):
        if getattr(msg, "type", None) == "ai":
            return str(getattr(msg, "content", "") or "")
    return ""


def _load(raw: str) -> dict:
    text = raw.strip()
    fenced = re.search(r"```(?:json)?\s*([\s\S]*?)```", text)
    if fenced:
        text = fenced.group(1).strip()
    m = _JSON.search(text)
    if not m:
        return {}
    try:
        data = json.loads(m.group(0))
    except json.JSONDecodeError:
        return {}
    return data if isinstance(data, dict) else {}
