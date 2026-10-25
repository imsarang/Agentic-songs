"""One terminal: live graphs from analyst:live:global. Never exits on silence."""

from __future__ import annotations

import os
import shutil

import redis

LIVE_KEY = os.getenv("LIVE_KEY", "analyst:live:global")
REDIS_URL = os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0")
WIDTH = 60
BLOCKS = "▁▂▃▄▅▆▇█"

# name, field, low, high
GRAPHS = (
    ("loudness dB", "loudness_db", -60.0, 0.0),
    ("energy", "energy", 0.0, 1.0),
    ("bass", "bass", 0.0, 1.0),
    ("mid", "mid", 0.0, 1.0),
    ("high", "high", 0.0, 1.0),
    ("brightness", "brightness", 0.0, 1.0),
    ("vibe", "vibe", 0.0, 1.0),
    ("beat hits", "beat_hits", 0.0, 8.0),
)


def _num(raw: str) -> float | None:
    if raw is None or raw == "":
        return None
    try:
        return float(raw)
    except ValueError:
        return None


def _silent(fields: dict) -> bool:
    return str(fields.get("silence", "")).lower() == "true"


def _spark(values: list[float], lo: float, hi: float) -> str:
    if not values:
        return "·" * WIDTH
    chunk = values[-WIDTH:]
    span = hi - lo or 1.0
    out = []
    for v in chunk:
        t = min(1.0, max(0.0, (float(v) - lo) / span))
        out.append(BLOCKS[min(len(BLOCKS) - 1, int(t * len(BLOCKS)))])
    return "".join(out).rjust(WIDTH, "·")


def _paint(series: list[dict], status: str, latest: dict | None) -> None:
    cols = shutil.get_terminal_size((80, 24)).columns
    lines = [f"shanks live    {status}"]
    if latest:
        lines.append(
            f"mood {latest.get('mood', '-')}   "
            f"vibe {latest.get('vibe', '-')}   "
            f"tempo {latest.get('tempo_bpm') or '-'} bpm   "
            f"hits {latest.get('beat_hits', '-')}"
        )
    else:
        lines.append("mood -   vibe -   tempo -   hits -")
    lines.append("")
    for label, field, lo, hi in GRAPHS:
        vals = [row[field] for row in series if isinstance(row.get(field), (int, float))]
        last = f"{vals[-1]:.2f}" if vals else "-"
        lines.append(f"{label:<12} {last:>7}  {_spark(vals, lo, hi)}")
    frame = "\n".join(lines)
    print("\033[H\033[J" + frame[: cols * 40], flush=True)


def _absorb(fields: dict, series: list[dict], status: str) -> tuple[list[dict], str, dict]:
    sample = {}
    for _, field, _, _ in GRAPHS:
        n = _num(fields.get(field, ""))
        if n is not None:
            sample[field] = n
    sample["mood"] = str(fields.get("mood") or "-")
    sample["tempo_bpm"] = _num(fields.get("tempo_bpm", ""))
    if _silent(fields):
        label = "silence    previous track" if series else "silence"
        return series, label, sample
    if status.startswith("silence"):
        series = []
    series.append(sample)
    if len(series) > 1800:
        del series[:-1800]
    return series, "live", sample


def main() -> None:
    r = redis.Redis.from_url(REDIS_URL, decode_responses=True)
    series: list[dict] = []
    status = "waiting"
    latest: dict | None = None
    last_id = "0-0"

    history = list(reversed(r.xrevrange(LIVE_KEY, count=WIDTH)))
    for entry_id, fields in history:
        series, status, latest = _absorb(fields, series, status)
        last_id = entry_id
    if not history:
        last_id = "$"

    _paint(series, status, latest)
    while True:
        batch = r.xread({LIVE_KEY: last_id}, block=2000, count=20)
        if not batch:
            _paint(series, status, latest)
            continue
        for _stream, entries in batch:
            for entry_id, fields in entries:
                last_id = entry_id
                series, status, latest = _absorb(fields, series, status)
        _paint(series, status, latest)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nclosed", flush=True)