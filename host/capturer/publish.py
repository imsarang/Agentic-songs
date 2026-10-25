"""One-shot: capture → features → Redis. Invoked by sample.sh."""

from __future__ import annotations

import json
import os
import sys
import time

os.environ.setdefault("JACK_NO_START_SERVER", "1")

import redis

from capturer import capture, features

WIRE_KEY = os.getenv("WIRE_KEY", "analyst:wire:latest")
LIVE_KEY = os.getenv("LIVE_KEY", "analyst:live:global")
HITS_KEY = os.getenv("HITS_KEY", "analyst:live:hits")

def _tempo(r, feat: dict, now: float, seconds: float) -> float | None:
    start = now - seconds
    for off in feat.get("hit_offsets") or []:
        r.rpush(HITS_KEY, str(start + float(off)))
    r.ltrim(HITS_KEY, -64, -1)
    times = [float(x) for x in r.lrange(HITS_KEY, 0, -1)]
    return features.tempo_bpm(times, now)
    
def main() -> int:
    url = os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0")
    r = redis.Redis.from_url(url, decode_responses=True)
    pcm, device = capture.record_slice()
    now = time.time()
    feat = features.from_pcm(pcm, sr=capture.sample_rate())
    bpm = _tempo(r, feat, now, capture.window_seconds())
    feat["tempo_bpm"] = None if bpm is None else round(bpm, 1)
    feat.pop("hit_offsets", None)
    payload = {"ts": now, "features": feat, "device": device}
    r.set(WIRE_KEY, json.dumps(payload))
    r.xadd(
        LIVE_KEY,
        {k: str(v) for k, v in {
            "ts": payload["ts"],
            "loudness_db": feat["loudness_db"],
            "energy": feat["energy"],
            "peak_db": feat["peak_db"],
            "crest": feat["crest"],
            "bass": feat["bass"],
            "mid": feat["mid"],
            "high": feat["high"],
            "brightness": feat["brightness"],
            "beat_hits": feat["beat_hits"],
            "tempo_bpm": "" if bpm is None else feat["tempo_bpm"],
            "vibe": feat["vibe"],
            "mood": feat["mood"],
            "silence": bool(feat["silence"]),
        }.items()},
        maxlen=1800,
        approximate=True,
    )
    print(
        f"wire ok loudness={feat.get('loudness_db')} "
        f"energy={feat.get('energy')} silence={feat.get('silence')} "
        f"target={device.get('target')}",
        flush=True,
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as e:
        print(f"wire error: {e}", file=sys.stderr, flush=True)
        raise SystemExit(1)
