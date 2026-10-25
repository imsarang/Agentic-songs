from __future__ import annotations

import json
import re

from langchain_core.messages import HumanMessage, SystemMessage

from shanks_agent.graph.prompts.base import BASE
from shanks_agent.graph.prompts.machine import MACHINE
from shanks_agent.graph.prompts.router import ROUTER
from shanks_agent.graph.prompts.spotify import SPOTIFY
from shanks_agent.llm.failover import FailoverLLM

DOMAINS = ("spotify", "machine")

_DOMAIN_PROMPTS = {
    "spotify": SPOTIFY,
    "machine": MACHINE,
}


def build_system_prompt(domain: str, tools: tuple[str, ...]) -> str:
    tools_blob = "\n".join(f"- {n}" for n in tools)
    body = _DOMAIN_PROMPTS.get(domain, SPOTIFY)
    return BASE.format(tools=tools_blob) + "\n" + body


def _parse_domain(raw: str) -> str | None:
    text = raw.strip()
    m = re.search(r"\{[\s\S]*\}", text)
    if m:
        text = m.group(0)
    try:
        data = json.loads(text)
    except json.JSONDecodeError:
        return None
    domain = str(data.get("domain", "")).lower().strip()
    return domain if domain in DOMAINS else None


async def resolve_domain(
    llm: FailoverLLM,
    text: str,
    domain: str | None = None,
) -> str:
    if domain in DOMAINS:
        return domain
    ai = await llm.acomplete(
        [SystemMessage(content=ROUTER), HumanMessage(content=text)]
    )
    if isinstance(ai.content, list):
        raw = "".join(
            b.get("text", "")
            for b in ai.content
            if isinstance(b, dict) and b.get("type") == "text"
        )
    else:
        raw = str(ai.content or "")
    return _parse_domain(raw) or "spotify"
