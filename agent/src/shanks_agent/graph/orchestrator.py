import json
import re

from langchain_core.messages import HumanMessage, SystemMessage

from shanks_agent.llm.failover import FailoverLLM

DOMAINS = ("spotify", "machine", "google", "creative")

_PLAN = """Pick which specialists to run for this user message.
Reply ONLY JSON: {"domains":["spotify"]} or {"domains":["creative","google"]}
Use more than one only when the user clearly asks for two different things.
Keep the list in the order they should run.
Recent lines are one line per domain, not the full history. Choose domains for the new message.
A revision continues the domain that already holds that piece. A new request does not.
spotify = music search, play, pause, skip, playlists.
machine = this computer: specs, CPU, GPU, memory, disk, load, processes, battery, network, logs, health.
google = gmail, contacts, drafts, send mail, trash mail, drive space. Reads and sends. Does not write the prose.
creative = writes the words: a mail body, poem, note, caption, or story.
When the user asks to write a letter or mail, the runner always does creative then google.
You can return that pair. Sending a letter Creative already wrote is google alone. Do not search the mailbox for it.
google alone to read, search, trash, or check drive.
creative alone when nothing has to be sent or played.
"""


async def plan(llm: FailoverLLM, text: str, hint: str = "", forced: str | None = None) -> list[str]:
    if forced in DOMAINS:
        return [forced]
    msgs = [SystemMessage(content=_PLAN)]
    if hint:
        msgs.append(HumanMessage(content="Recent turns:\n" + hint))
    msgs.append(HumanMessage(content=text))
    ai = await llm.acomplete(msgs)
    raw = ai.content
    if isinstance(raw, list):
        raw = "".join(b.get("text", "") for b in raw if isinstance(b, dict) and b.get("type") == "text")
    raw = str(raw or "")
    m = re.search(r"\{[\s\S]*\}", raw)
    try:
        data = json.loads(m.group(0) if m else raw)
        out = [d for d in data.get("domains", []) if d in DOMAINS]
        if out:
            return out
    except Exception:
        pass
    return ["spotify"]


async def merge(llm: FailoverLLM, user: str, stubs: list[str]) -> str:
    if len(stubs) == 1:
        return stubs[0]
    ai = await llm.acomplete([
        SystemMessage(content="Combine these specialist results into one short reply. No tools."),
        HumanMessage(content=user + "\n\n" + "\n---\n".join(stubs)),
    ])
    c = ai.content
    if isinstance(c, list):
        return "".join(b.get("text", "") for b in c if isinstance(b, dict) and b.get("type") == "text")
    return str(c or "")
