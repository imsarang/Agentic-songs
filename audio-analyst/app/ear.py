"""Read the latest wire features from Redis (written by ./host daemon)."""

from __future__ import annotations

import os
import time

from app import store

WIRE_KEY = os.getenv("WIRE_KEY", "analyst:wire:latest")
WIRE_MAX_AGE_SEC = float(os.getenv("WIRE_MAX_AGE_SEC", "30"))


def sample_features() -> tuple[dict, dict]:
    """Return (features, device_meta) from analyst:wire:latest."""
    raw = store.r.get(WIRE_KEY)
    if not raw:
        raise RuntimeError(
            f"no wire data at {WIRE_KEY}. Start the host daemon: ./host/run.sh"
        )
    data = __import__("json").loads(raw)
    ts = float(data.get("ts") or 0)
    age = time.time() - ts
    if ts <= 0 or age > WIRE_MAX_AGE_SEC:
        raise RuntimeError(
            f"wire data stale ({age:.0f}s old). Is ./host/run.sh running?"
        )
    feat = data.get("features") or {}
    device = data.get("device") or {"device": "host-wire", "name": WIRE_KEY}
    return feat, device
