import os
import time

# for now storing in-memory
# key : tokens,last time stamp
_buckets: dict[str, tuple[float, float]] = {}

def check(key: str) -> str | None:
    rate = float(os.getenv("AGENT_RATE_PER_MIN", "30"))

    if rate <=0: return None

    now = time.monotonic()

    tokens, last = _buckets.get(key, (rate, now))

    tokens = min(rate, tokens + (now-last)*(rate/60.0))

    if tokens < 1:
        _buckets[key] = (tokens, now)
        return "rate limited"
    _buckets[key] = (tokens - 1, now)
    return None