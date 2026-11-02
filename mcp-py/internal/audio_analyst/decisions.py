"""Phase 3: rules on a feature vector → structured decisions + bullets."""

from __future__ import annotations

import os

# Tunable floors (loudness is dBFS; energy is 0..1-ish).
LOUD_DB = float(os.getenv("RULE_TOO_LOUD_DB", "-12"))
HIGH_ENERGY = float(os.getenv("RULE_HIGH_ENERGY", "0.55"))
LOW_ENERGY = float(os.getenv("RULE_LOW_ENERGY", "0.20"))


def from_features(feat: dict) -> tuple[list[dict], list[str]]:
    """Return (decisions, summary_bullets) for one feature vector."""
    if not feat:
        return (
            [{"type": "suggest", "value": "no_features"}],
            ["No features available"],
        )

    decisions: list[dict] = []
    bullets: list[str] = []

    silence = bool(feat.get("silence"))
    loud = float(feat.get("loudness_db", -90))
    energy = float(feat.get("energy", 0.0))

    if silence:
        decisions.append({"type": "suggest", "value": "check_audio_source"})
        bullets.append("Silence / nothing useful on the wire")
        return decisions, bullets

    bullets.append(f"Loudness {loud:.1f} dB")
    bullets.append(f"Energy {energy:.2f}")

    if loud >= LOUD_DB:
        decisions.append({"type": "flag", "value": "too_loud"})
        decisions.append({"type": "suggest", "value": "lower_volume"})
        bullets.append("Quite loud — consider lowering volume")
    else:
        decisions.append({"type": "suggest", "value": "keep_playing"})

    if energy >= HIGH_ENERGY:
        decisions.append({"type": "mood", "value": "hype"})
        bullets.append("High energy — hype")
    elif energy <= LOW_ENERGY:
        decisions.append({"type": "mood", "value": "chill"})
        bullets.append("Low energy — chill")
    else:
        decisions.append({"type": "mood", "value": "mid"})
        bullets.append("Mid energy")

    if feat.get("samples_used") is not None:
        bullets.append(f"Averaged over {int(feat['samples_used'])} non-silent samples")

    return decisions, bullets


def attach(payload: dict) -> dict:
    """Mutate/return payload with decisions + summary_bullets from its features."""
    feat = payload.get("features")
    if isinstance(feat, dict):
        dec, bullets = from_features(feat)
        payload["decisions"] = dec
        payload["summary_bullets"] = bullets
    return payload
