"""RedisTimeSeries samples, plus the hit list used for tempo."""

import os

from capturer import features

RETENTION_MS = 30 * 60 * 1000
HITS_KEY = os.getenv("HITS_KEY", "analyst:live:hits")

GROUPS = {
    "level": ("loudness_db", "peak_db"),
    "bands": ("bass", "mid", "high"),
    "tone": ("energy", "brightness"),
    "crest": ("crest",),
    "hits": ("beat_hits",),
    "vibe": ("vibe",),
    "tempo": ("tempo_bpm",),
    "mood": ("mood_code",),
}

MOOD_CODE = {"silence": 0, "chill": 1, "mid": 2, "hype": 3}


def tempo(r, offsets, now: float, seconds: float) -> float | None:
    if offsets:
        start = now - seconds
        r.rpush(HITS_KEY, *[str(start + float(off)) for off in offsets])
        r.ltrim(HITS_KEY, -64, -1)
    times = [float(x) for x in r.lrange(HITS_KEY, 0, -1)]
    bpm = features.tempo_bpm(times, now)
    return None if bpm is None else round(bpm, 1)


def _rows(feat: dict, ts_ms: int):
    sample = {**feat, "mood_code": MOOD_CODE.get(str(feat.get("mood")), 0)}
    for group, names in GROUPS.items():
        for name in names:
            value = sample.get(name)
            if value is None or value == "":
                continue
            yield f"live:{name}", ts_ms, float(value), ("group", group, "name", name)


def write(r, feat: dict, ts_ms: int) -> None:
    rows = list(_rows(feat, ts_ms))
    if not rows:
        return
    probe = r.pipeline(transaction=False)
    for key, *_ in rows:
        probe.exists(key)
    pipe = r.pipeline(transaction=False)
    for (key, ts, value, labels), found in zip(rows, probe.execute()):
        if not found:
            pipe.execute_command(
                "TS.CREATE", key,
                "RETENTION", RETENTION_MS,
                "DUPLICATE_POLICY", "LAST",
                "LABELS", *labels,
            )
        pipe.execute_command("TS.ADD", key, ts, value)
    pipe.execute()
