"""One-shot: capture → features → Redis. Invoked by sample.sh."""

import json
import os
import sys
import time

os.environ.setdefault("JACK_NO_START_SERVER", "1")

import redis

from capturer import capture, features, series

WIRE_KEY = os.getenv("WIRE_KEY", "analyst:wire:latest")


def main() -> int:
    r = redis.Redis.from_url(
        os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0"),
        decode_responses=True,
    )
    pcm, device = capture.record_slice()
    now = time.time()
    feat = features.from_pcm(pcm, sr=capture.sample_rate())
    feat["tempo_bpm"] = series.tempo(r, feat.pop("hit_offsets", None), now, capture.window_seconds())
    r.set(WIRE_KEY, json.dumps({"ts": now, "features": feat, "device": device}))
    series.write(r, feat, int(now * 1000))
    print(
        f"wire ok loudness={feat['loudness_db']} energy={feat['energy']} "
        f"silence={feat['silence']} target={device['target']}",
        flush=True,
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as e:
        print(f"wire error: {e}", file=sys.stderr, flush=True)
        raise SystemExit(1)
