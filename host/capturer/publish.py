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


def main() -> int:
    url = os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0")
    r = redis.Redis.from_url(url, decode_responses=True)

    pcm, device = capture.record_slice()
    feat = features.from_pcm(pcm)
    payload = {"ts": time.time(), "features": feat, "device": device}
    r.set(WIRE_KEY, json.dumps(payload))
    r.xadd(
        LIVE_KEY,
        {
            "ts": str(payload["ts"]),
            "loudness_db": str(feat.get("loudness_db")),
            "energy": str(feat.get("energy")),
            "silence": str(bool(feat.get("silence"))),
        },
        maxlen=500,
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
