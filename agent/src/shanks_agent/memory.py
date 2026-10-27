import json
import os
import re

from redis.asyncio import Redis

r = Redis.from_url(os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0"), decode_responses=True)

DOMAINS = ("spotify", "machine", "google", "creative")
KEEP = 5
_DRAFT = re.compile(r"draft_id=(\S+)")
_FOLLOW = {
    "yes", "y", "yeah", "yep", "ok", "okay", "sure",
    "no", "send", "send it", "do it", "go ahead",
    "delete", "delete it", "delete them", "all",
}

def _key(thread_id: str, domain: str) -> str:
    return f"shanks:thread:{thread_id}:{domain}"

def _last_key(thread_id: str) -> str:
    return f"shanks:thread:{thread_id}:last"

async def load(thread_id: str, domain: str) -> list[dict]:
    raw = await r.lrange(_key(thread_id, domain), -KEEP, -1)
    return [json.loads(item) for item in raw]

async def index(thread_id: str) -> str:
    lines = []
    for domain in DOMAINS:
        rows = await load(thread_id, domain)
        if rows:
            lines.append(_index_line(domain, rows[-1]))
    return "\n".join(lines)

def latest_note(rows: list[dict]) -> str:
    if not rows:
        return ""
    return str(rows[-1].get("note") or "").strip()

def history(rows: list[dict]) -> str:
    blocks = []
    for row in rows:
        text = f"user={row.get('user', '')}\n{row.get('note', '')}".strip()
        if text:
            blocks.append(text)
    return "\n---\n".join(blocks)

def continues(text: str) -> bool:
    cleaned = " ".join(text.lower().split()).strip(".!?")
    return cleaned in _FOLLOW or _is_handle(cleaned)

async def last_domain(thread_id: str) -> str | None:
    value = await r.get(_last_key(thread_id))
    return value if value in DOMAINS else None

def note_from(messages) -> str:
    tools: list[str] = []
    letter = ""
    for msg in messages or []:
        text = str(getattr(msg, "content", "") or "").strip()
        if text.startswith("Creative draft"):
            letter = text.split("\n", 1)[-1].strip()
            continue
        if text.startswith("[tool "):
            tools.append(text)
    if not letter:
        for msg in reversed(messages or []):
            if getattr(msg, "type", None) != "ai":
                continue
            text = str(getattr(msg, "content", "") or "").strip()
            if text and not text.startswith('{"action"'):
                letter = text
                break
    # Tool rows hold search_id and playlist names. The finish text does not.
    if tools:
        blob = "\n---\n".join(tools)
        return blob[-4000:]
    return letter[:4000]

async def save(thread_id: str, domain: str, user: str, note: str) -> None:
    key = _key(thread_id, domain)
    await r.rpush(key, json.dumps({
        "user": user[:800],
        "note": note[:4000],
    }))
    await r.ltrim(key, -KEEP, -1)
    await r.set(_last_key(thread_id), domain)

def _index_line(domain: str, row: dict) -> str:
    user = _clip(row.get("user") or "", 80)
    note = row.get("note") or ""
    found = _DRAFT.search(note)
    tail = f"draft_id={found.group(1)}" if found else _clip(note, 100)
    line = f"{domain}: {user}"
    return f"{line} | {tail}" if tail else line

def _clip(text: str, limit: int) -> str:
    text = " ".join(text.split())
    return text if len(text) <= limit else text[:limit].rstrip() + "…"

def _is_handle(text: str) -> bool:
    return text.startswith("r-") and any(c.isdigit() for c in text) and " " not in text
