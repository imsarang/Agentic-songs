"""PCM → loudness_db / energy (RMS). Waveform is not persisted."""

from __future__ import annotations

import numpy as np


def loudness_db(pcm: np.ndarray) -> float:
    rms = float(np.sqrt(np.mean(np.square(pcm)) + 1e-12))
    return 20.0 * float(np.log10(rms))


def energy(pcm: np.ndarray) -> float:
    rms = float(np.sqrt(np.mean(np.square(pcm)) + 1e-12))
    return float(min(1.0, rms * 10.0))


def from_pcm(pcm: np.ndarray) -> dict:
    db = loudness_db(pcm)
    return {
        "loudness_db": round(db, 2),
        "energy": round(energy(pcm), 3),
        "samples": int(pcm.size),
        "silence": db < -50.0,
    }
