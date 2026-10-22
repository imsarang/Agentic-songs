"""decide(): Redis samples + rules. Features come from host via app.ear."""

from __future__ import annotations

import time

from app import decisions, features, store
from app.ear import sample_features


def _with_rules(payload: dict) -> dict:
    return decisions.attach(payload)


def _take_sample() -> tuple[dict, dict] | dict:
    try:
        return sample_features()
    except Exception as e:
        return {"ok": False, "error": f"capture failed: {e}"}


def _finalize(tid: str, samples: list[dict], source: str = "live") -> dict:
    n = len(samples)
    if n == 0:
        return {"ok": False, "error": "no samples to finalize", "trackId": tid}

    weights = store.WEIGHTS[:n]
    final_feat = features.weighted_mean(samples, weights)
    payload = {
        "trackId": tid,
        "features": final_feat,
        "sample_count": n,
        "weights": weights,
        "finalized": True,
    }
    store.save_final(tid, payload)
    return _with_rules({"ok": True, "status": "done", "source": source, **payload})


def decide(track_id: str, finalize: bool = False) -> dict:
    tid = store.normalize_id(track_id)
    if not tid:
        return {"ok": False, "error": "trackId required"}

    cached = store.get_final(tid)
    if cached:
        return _with_rules({"ok": True, "status": "done", "source": "redis", **cached})

    samples = store.get_partial(tid)
    i = len(samples)

    if finalize:
        if i == 0:
            got = _take_sample()
            if isinstance(got, dict) and got.get("ok") is False:
                return got
            feat, device = got  # type: ignore[misc]
            store.save_partial_sample(
                tid, 1, feat, {"last_ts": time.time(), "device": device, "i": 1}
            )
            samples = [feat]
        return _finalize(tid, samples)

    if i >= store.SAMPLE_COUNT:
        last = samples[-1]
        return _with_rules(
            {
                "ok": True,
                "status": "full",
                "i": i,
                "n": store.SAMPLE_COUNT,
                "features": last,
                "trackId": tid,
                "hint": "sample cap reached; call with finalize=true for weighted avg",
            }
        )

    ok, wait_ms = store.gap_ok(tid)
    if not ok:
        last = samples[-1] if samples else None
        out: dict = {
            "ok": True,
            "status": "waiting",
            "i": i,
            "n": store.SAMPLE_COUNT,
            "wait_ms": wait_ms,
            "trackId": tid,
        }
        if last is not None:
            out["features"] = last
            out["reused"] = True
            return _with_rules(out)
        return out

    got = _take_sample()
    if isinstance(got, dict) and got.get("ok") is False:
        return got
    feat, device = got  # type: ignore[misc]

    next_i = i + 1
    meta = {"last_ts": time.time(), "device": device, "i": next_i}
    store.save_partial_sample(tid, next_i, feat, meta)

    return _with_rules(
        {
            "ok": True,
            "status": "sample",
            "i": next_i,
            "n": store.SAMPLE_COUNT,
            "partial_key": f"{tid}-{next_i}",
            "features": feat,
            "device": device,
            "trackId": tid,
        }
    )
