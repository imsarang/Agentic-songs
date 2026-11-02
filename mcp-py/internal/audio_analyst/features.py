"""Aggregate stored feature dicts (no PCM — host capturer owns waveform math)."""

from __future__ import annotations


def weighted_mean(samples: list[dict], weights: list[float]) -> dict:
    """Avg loudness_db + energy; silence samples get weight 0."""
    nums = {"loudness_db": 0.0, "energy": 0.0}
    wsum = 0.0
    used = 0.0

    for i, s in enumerate(samples):
        w = weights[i] if i < len(weights) else 1.0
        if s.get("silence"):
            w = 0.0
        if w <= 0:
            continue
        nums["loudness_db"] += w * float(s["loudness_db"])
        nums["energy"] += w * float(s["energy"])
        wsum += w
        used += 1

    if wsum <= 0:
        return {
            "loudness_db": -90.0,
            "energy": 0.0,
            "silence": True,
            "samples_used": 0,
        }
    return {
        "loudness_db": round(nums["loudness_db"] / wsum, 2),
        "energy": round(nums["energy"] / wsum, 3),
        "silence": False,
        "samples_used": used,
    }
