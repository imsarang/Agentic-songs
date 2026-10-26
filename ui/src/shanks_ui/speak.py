"""Speak a reply in the reference voice. Lower Spotify while it plays."""

import os
import re
import subprocess
import textwrap
from pathlib import Path

os.environ.setdefault("PYTORCH_CUDA_ALLOC_CONF", "expandable_segments:True")
os.environ.setdefault("COQUI_TOS_AGREED", "1")

import numpy as np
import torch
import queue
import threading
from TTS.api import TTS

engine = None
CLIP = Path.home() / ".local/share/voices/reference.wav"
UNITS = {"GB": "gigabytes", "MB": "megabytes", "TB": "terabytes", "KB": "kilobytes"}


def load() -> None:
    global engine
    if engine is not None:
        return
    engine = TTS("tts_models/multilingual/multi-dataset/xtts_v2")
    if torch.cuda.is_available():
        engine.synthesizer.tts_model.cuda()
        engine.synthesizer.use_cuda = True


def _output(args: list[str]) -> str:
    return subprocess.check_output(args, text=True, stderr=subprocess.DEVNULL)


def spotify_now() -> tuple[str, str] | None:
    """Playing Spotify stream id and its volume, if music is on."""
    status = _output(["wpctl", "status"])
    found = re.search(r"^\s+(\d+)\. Spotify\s*$", status, re.M)
    if not found or "[active]" not in status[found.end():].split("\n\n", 1)[0]:
        return None
    node = found.group(1)
    level = re.search(r"([0-9.]+)", _output(["wpctl", "get-volume", node]))
    return (node, level.group(1)) if level else None


def set_volume(node: str, level: str) -> None:
    subprocess.run(["wpctl", "set-volume", node, level], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def lines_to_say(text: str) -> list[str]:
    """Strip symbols, say units as words, and keep each line short for a 4 GB GPU."""
    text = re.sub(r"\b(GB|MB|TB|KB)\b", lambda match: UNITS[match.group(1)], text or "")
    text = re.sub(r"(\d)\.(\d)", r"\1 point \2", text)
    text = re.sub(r"[-–—*#_`~^|<>{}\[\]():=+&%@/\\]+", " ", text)
    return textwrap.wrap(text, width=80, break_long_words=False)


def speak(text: str) -> None:
    lines = lines_to_say(text)
    clip = Path(os.getenv("VOICE_WAV", CLIP))
    if not lines:
        return
    if not clip.is_file():
        raise FileNotFoundError(f"voice sample missing: {clip}")

    load()
    model = engine.synthesizer.tts_model
    rate = int(engine.synthesizer.output_sample_rate or 24000)
    voice, tone = model.get_conditioning_latents(audio_path=str(clip))
    saved = spotify_now()
    if saved:
        set_volume(saved[0], "0.20")

    player = subprocess.Popen(
        ["pw-cat", "-p", "--format", "s16", "--rate", str(rate), "--channels", "1", "-"],
        stdin=subprocess.PIPE,
    )
    pending = queue.Queue(maxsize=3)

    def produce() -> None:
        try:
            for line in lines:
                for audio in model.inference_stream(line, "en", voice, tone, stream_chunk_size=20):
                    samples = audio.detach().float().cpu().numpy().reshape(-1).clip(-1, 1)
                    # 1.5x by keeping every sample along a shorter timeline. Same play rate as before.
                    picked = samples[np.arange(0, len(samples), 1.5).astype(int)]
                    pending.put((picked * 32767).astype(np.int16).tobytes())
                torch.cuda.empty_cache()
            pending.put(None)
        except Exception as exc:
            pending.put(exc)

    threading.Thread(target=produce, daemon=True).start()

    try:
        while True:
            chunk = pending.get()
            if chunk is None:
                break
            if isinstance(chunk, Exception):
                raise chunk
            player.stdin.write(chunk)
            player.stdin.flush()
        player.stdin.close()
        player.wait()
    finally:
        if player.poll() is None:
            try:
                player.stdin.close()
            except Exception:
                pass
            try:
                player.wait(timeout=5)
            except subprocess.TimeoutExpired:
                player.kill()
        if saved:
            set_volume(*saved)
