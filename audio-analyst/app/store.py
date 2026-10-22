"""Redis: partial samples as {trackId}-1..N, final as {trackId}:final."""

from __future__ import annotations

import json
import os
import time

import redis

r = redis.Redis.from_url(
    os.getenv("REDIS_URL", "redis://127.0.0.1:6379/0"),
    decode_responses=True,
)

SAMPLE_COUNT = int(os.getenv("SAMPLE_COUNT", "5"))
SAMPLE_GAP_MS = int(os.getenv("SAMPLE_GAP_MS", "15000"))
WEIGHTS = [float(x) for x in os.getenv("SAMPLE_WEIGHTS", "1,2,3,4,5").split(",")]
FINAL_TTL = int(os.getenv("FINAL_TTL_SEC", "2592000"))  # 30d; 0 = no expiry


def normalize_id(track_id: str) -> str:
    """Bare Spotify track id (strip spotify:track: if someone pastes a URI)."""
    u = track_id.strip()
    if u.startswith("spotify:track:"):
        return u.split(":")[-1]
    return u


def final_key(track_id: str) -> str:
    return f"analyst:track:{normalize_id(track_id)}:final"


def partial_key(track_id: str, index: int) -> str:
    # e.g. analyst:track:5lWSa1rmuSL6OBPOnkAqoa-1
    return f"analyst:track:{normalize_id(track_id)}-{index}"


def meta_key(track_id: str) -> str:
    return f"analyst:track:{normalize_id(track_id)}:meta"


def get_final(track_id: str) -> dict | None:
    raw = r.get(final_key(track_id))
    return json.loads(raw) if raw else None


def get_partial(track_id: str) -> list[dict]:
    """Load samples 1..SAMPLE_COUNT that exist (hyphen keys)."""
    out: list[dict] = []
    for i in range(1, SAMPLE_COUNT + 1):
        raw = r.get(partial_key(track_id, i))
        if not raw:
            break
        out.append(json.loads(raw))
    return out


def partial_count(track_id: str) -> int:
    return len(get_partial(track_id))


def get_meta(track_id: str) -> dict:
    raw = r.get(meta_key(track_id))
    return json.loads(raw) if raw else {}


def save_partial_sample(track_id: str, index: int, sample: dict, meta: dict) -> None:
    """Write one slice as {trackId}-{index} (1-based)."""
    r.set(partial_key(track_id, index), json.dumps(sample))
    r.set(meta_key(track_id), json.dumps(meta))


def save_final(track_id: str, payload: dict) -> None:
    tid = normalize_id(track_id)
    key = final_key(tid)
    r.set(key, json.dumps(payload))
    if FINAL_TTL > 0:
        r.expire(key, FINAL_TTL)
    # drop partials + meta
    keys = [partial_key(tid, i) for i in range(1, SAMPLE_COUNT + 1)]
    keys.append(meta_key(tid))
    r.delete(*keys)


def gap_ok(track_id: str) -> tuple[bool, int]:
    """True if enough time since last sample. Returns (ok, wait_ms)."""
    meta = get_meta(track_id)
    last = float(meta.get("last_ts", 0))
    elapsed_ms = (time.time() - last) * 1000
    if last == 0 or elapsed_ms >= SAMPLE_GAP_MS:
        return True, 0
    return False, int(SAMPLE_GAP_MS - elapsed_ms)
