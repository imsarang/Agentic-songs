"""PCM → loudness_db / energy (RMS). Waveform is not persisted."""

from __future__ import annotations

import numpy as np

_BASS = (20.0, 150.0)
_MID = (150.0, 2000.0)
_HIGH = (2000.0, 8000.0)

def _rms(pcm: np.ndarray) -> float:
    return float(np.sqrt(np.mean(np.square(pcm)) + 1e-12))

def loudness_db(pcm: np.ndarray) -> float:
    return 20.0 * float(np.log10(_rms(pcm)))

def energy(pcm: np.ndarray) -> float:
    rms = float(np.sqrt(np.mean(np.square(pcm)) + 1e-12))
    return float(min(1.0, rms * 10.0))

def peak_db(pcm: np.ndarray) -> float:
    peak = float(np.max(np.abs(pcm))) + 1e-12
    return 20.0 * float(np.log10(peak))

def crest(pcm: np.ndarray) -> float:
    rms = _rms(pcm)
    if rms < 1e-8:
        return 0.0
    return float(np.max(np.abs(pcm)) / rms)

def _power(pcm: np.ndarray, sr: int) -> tuple[np.ndarray, np.ndarray]:
    spec = np.abs(np.fft.rfft(pcm))
    freqs = np.fft.rfftfreq(pcm.size, d=1.0 / sr)
    return freqs, spec ** 2

def band_share(pcm: np.ndarray, sr: int, lo: float, hi: float) -> float:
    freqs, power = _power(pcm, sr)
    total = float(np.sum(power)) + 1e-12
    mask = (freqs >= lo) & (freqs < hi)
    return float(np.sum(power[mask]) / total)

def brightness(pcm: np.ndarray, sr: int) -> float:
    """0-1. 1 means the spectral centroid is up around sr/4."""
    spec = np.abs(np.fft.rfft(pcm))
    freqs = np.fft.rfftfreq(pcm.size, d=1.0 / sr)
    centroid = float(np.sum(freqs * spec) / (float(np.sum(spec)) + 1e-12))
    return float(min(1.0, centroid / (sr / 4.0)))

def hit_offsets(pcm: np.ndarray, sr: int, frame_ms: int = 50) -> list[float]:
    """Seconds from the start of this slice where the level jumps."""
    n = max(1, int(sr * frame_ms / 1000))
    rms: list[float] = []
    for i in range(0, pcm.size - n + 1, n):
        rms.append(_rms(pcm[i : i + n]))
    offsets: list[float] = []
    for i in range(1, len(rms)):
        if rms[i] > rms[i - 1] * 1.6 and rms[i] > 0.02:
            offsets.append(i * frame_ms / 1000.0)
    return offsets

def tempo_bpm(hit_times: list[float], now: float, horizon: float = 15.0) -> float | None:
    times = [t for t in hit_times if now - t <= horizon]
    if len(times) < 4:
        return None
    gaps = np.diff(np.array(times, dtype=np.float64))
    gaps = gaps[(gaps >= 60.0 / 180.0) & (gaps <= 60.0 / 60.0)]
    if gaps.size == 0:
        return None
    return float(60.0 / np.median(gaps))

def from_pcm(pcm: np.ndarray, sr: int = 44100) -> dict:
    db = loudness_db(pcm)
    nrg = energy(pcm)
    silent = db < -50.0
    bright = 0.0 if silent else brightness(pcm, sr)
    vibe = 0.0 if silent else float(min(1.0, 0.5 * nrg + 0.5 * bright))
    if silent:
        mood = "silence"
    elif vibe >= 0.55:
        mood = "hype"
    elif vibe <= 0.20:
        mood = "chill"
    else:
        mood = "mid"
    offsets = [] if silent else hit_offsets(pcm, sr)
    return {
        "loudness_db": round(db, 2),
        "energy": round(nrg, 3),
        "peak_db": round(peak_db(pcm), 2),
        "crest": round(crest(pcm), 3),
        "bass": round(band_share(pcm, sr, *_BASS), 3),
        "mid": round(band_share(pcm, sr, *_MID), 3),
        "high": round(band_share(pcm, sr, *_HIGH), 3),
        "brightness": round(bright, 3),
        "beat_hits": len(offsets),
        "hit_offsets": [round(x, 3) for x in offsets],
        "vibe": round(vibe, 3),
        "mood": mood,
        "samples": int(pcm.size),
        "silence": silent,
    }