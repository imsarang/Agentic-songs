import os
import time

from redis import Redis

RETENTION_MS = 24 * 60 * 60 * 1000
_redis = None


def _client() -> Redis:
    global _redis
    if _redis is None:
        _redis = Redis.from_url(os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0"))
    return _redis


def write(latency_ms: float, empty: bool, count: int) -> None:
    ts = int(time.time() * 1000)
    rows = (
        ("retrieve:latency_ms", latency_ms, ("group", "retrieve", "name", "latency_ms")),
        ("retrieve:empty", 1.0 if empty else 0.0, ("group", "retrieve", "name", "empty")),
        ("retrieve:chunks", float(count), ("group", "retrieve", "name", "chunks")),
    )
    try:
        client = _client()
        probe = client.pipeline(transaction=False)
        for key, *_ in rows:
            probe.exists(key)
        pipe = client.pipeline(transaction=False)
        for (key, value, labels), found in zip(rows, probe.execute()):
            if not found:
                pipe.execute_command(
                    "TS.CREATE", key,
                    "RETENTION", RETENTION_MS,
                    "DUPLICATE_POLICY", "LAST",
                    "LABELS", *labels,
                )
            pipe.execute_command("TS.ADD", key, ts, value)
        pipe.execute()
    except Exception:
        return