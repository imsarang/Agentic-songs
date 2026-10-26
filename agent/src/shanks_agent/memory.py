import json
import os

from redis.asyncio import Redis

r = Redis.from_url(os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0"), decode_responses=True)

def _key(thread_id: str) -> str:
    return f"shanks:thread:{thread_id}"

async def load(thread_id: str) -> list[dict]:
    raw = await r.lrange(_key(thread_id), -10, -1)
    return [json.loads(item) for item in raw]

_FOLLOW = {
    "yes", "y", "yeah", "yep", "ok", "okay", "sure",
    "no", "send", "send it", "do it", "go ahead",
    "delete", "delete it", "delete them", "all",
}

def domain_of(rows: list[dict], text: str) -> str | None:
    if not rows:
        return None
    cleaned = " ".join(text.lower().split()).strip(".!?")
    if cleaned in _FOLLOW or _is_handle(cleaned):
        return rows[-1]["domain"]
    return None

def _is_handle(text: str) -> bool:
    if " " in text or len(text) < 2:
        return False
    return any(ch.isdigit() for ch in text)

def prompt(rows: list[dict]) -> str:
    if not rows:
        return ""
    row = rows[-1]
    return f"{row['domain']} | {row['user']} | {row.get('note', '')}"

def note_from(messages) -> str:
    note = ""
    for msg in messages or []:
        text = str(getattr(msg, "content", "") or "")
        if not text.startswith("[tool "):
            continue
        head, _, body = text.partition("\n")
        ids = [bit for bit in body.split() if "=" in bit and len(bit) < 60]
        note = head if not ids else head + " " + " ".join(ids[:6])
    return note[:200]

async def save(thread_id: str, domain: str, user: str, note: str) -> None:
    key = _key(thread_id)
    await r.rpush(key, json.dumps({
        "domain": domain,
        "user": user[:200],
        "note": note[:200],
    }))
    await r.ltrim(key, -10, -1)