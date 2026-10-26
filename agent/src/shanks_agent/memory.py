import json
import os

from redis.asyncio import Redis

r = Redis.from_url(os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0"), decode_responses=True)

def _key(thread_id: str) -> str:
    return f"shanks:thread:{thread_id}"

async def load(thread_id: str) -> list[dict]:
    raw = await r.lrange(_key(thread_id), -10, -1)
    return [json.loads(item) for item in raw]

def domain_of(rows: list[dict], text: str) -> str | None:
    if rows and len(text) < 40:
        return rows[-1]["domain"]
    return None

def prompt(rows: list[dict]) -> str:
    if not rows: return ""
    domain = rows[-1]["domain"]
    same = [row for row in rows if row["domain"] == domain]
    return "\n".join(
        f"{row['domain']} | {row['user']} | {row.get('note', '')}" for row in same
    )

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