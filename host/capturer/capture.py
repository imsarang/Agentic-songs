"""Record a short PCM slice from the default playback sink (PipeWire)."""

from __future__ import annotations

import os
import re
import subprocess

import numpy as np


def window_seconds() -> float:
    return max(0.5, float(os.getenv("WINDOW_MS", "5000")) / 1000.0)


def sample_rate() -> int:
    return int(os.getenv("SAMPLE_RATE", "44100"))


def _resolve_sink_target() -> str:
    env = os.getenv("AUDIO_TARGET", "").strip()
    if env:
        return env

    try:
        out = subprocess.check_output(
            ["wpctl", "status"], text=True, stderr=subprocess.DEVNULL
        )
    except (FileNotFoundError, subprocess.CalledProcessError) as e:
        raise RuntimeError(
            "wpctl not available; set AUDIO_TARGET to a PipeWire sink id/name"
        ) from e

    in_audio = False
    in_sinks = False
    for line in out.splitlines():
        if line.strip() == "Audio":
            in_audio = True
            continue
        if in_audio and line.strip() == "Video":
            break
        if not in_audio:
            continue
        if "Sinks:" in line:
            in_sinks = True
            continue
        if in_sinks and "Sources:" in line:
            break
        if not in_sinks:
            continue
        if "*" not in line:
            continue
        m = re.search(r"(\d+)\.", line)
        if m:
            return m.group(1)

    raise RuntimeError(
        "no default PipeWire sink found; set AUDIO_TARGET or check `wpctl status`"
    )


def record_slice() -> tuple[np.ndarray, dict]:
    sr = sample_rate()
    seconds = window_seconds()
    target = _resolve_sink_target()
    frames = int(sr * seconds)
    nbytes = frames * 4

    cmd = [
        "pw-record",
        f"--target={target}",
        f"--rate={str(sr)}",
        "--channels=1",
        "--format=f32",
        "--latency=50ms",
        "-",
    ]
    try:
        proc = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
    except FileNotFoundError as e:
        raise RuntimeError("pw-record not found (install pipewire-bin)") from e

    assert proc.stdout is not None
    chunks: list[bytes] = []
    got = 0
    try:
        while got < nbytes:
            chunk = proc.stdout.read(min(65536, nbytes - got))
            if not chunk:
                break
            chunks.append(chunk)
            got += len(chunk)
    finally:
        proc.kill()
        try:
            proc.wait(timeout=2)
        except subprocess.TimeoutExpired:
            proc.terminate()

    raw = b"".join(chunks)
    if len(raw) < 4:
        err = (proc.stderr.read() if proc.stderr else b"").decode(errors="replace")
        raise RuntimeError(
            f"pw-record produced no audio (target={target}). {err.strip()}"
        )

    n = (len(raw) // 4) * 4
    pcm = np.frombuffer(raw[:n], dtype="<f4").astype(np.float32, copy=True)
    np.clip(pcm, -1.0, 1.0, out=pcm)

    if pcm.size < max(1, frames // 4):
        raise RuntimeError(
            f"pw-record too short ({pcm.size} samples, want ~{frames}, target={target})"
        )

    device_meta = {
        "device": "pw-record",
        "target": target,
        "name": f"sink:{target}",
        "max_input_channels": 1,
        "default_samplerate": float(sr),
    }
    return pcm, device_meta
