"""Hold Left Ctrl, record the mic, transcribe with whisper-cli."""

import os
import re
import select
import signal
import subprocess
import tempfile
import threading
import time
from pathlib import Path

KEY_LEFTCTRL = 29
EV_KEY = 1
MIN_HOLD_S = 0.35
_BLANK = {"[BLANK_AUDIO]", "(silence)"}


def _wpctl(*args: str) -> str:
    return subprocess.check_output(["wpctl", *args], text=True, stderr=subprocess.DEVNULL)


def mic_target() -> tuple[str, str]:
    """pw-record --target wants the source serial, not the wpctl id."""
    env = os.getenv("AUDIO_SOURCE", "").strip()
    if env and not env.isdigit():
        return env, env
    node = env or _starred_source(_wpctl("status"))
    out = _wpctl("inspect", node)
    serial = re.search(r'object\.serial\s*=\s*"(\d+)"', out)
    desc = re.search(r'node\.description\s*=\s*"([^"]+)"', out)
    if not serial:
        raise RuntimeError(f"node {node} has no object.serial")
    return serial.group(1), (desc.group(1) if desc else node)


def _starred_source(status: str) -> str:
    audio = status.split("\nVideo\n", 1)[0]
    block = audio.split("Sources:", 1)[-1].split("Source endpoints:", 1)[0]
    for line in block.splitlines():
        if "*" in line and (match := re.search(r"(\d+)\.", line)):
            return match.group(1)
    raise RuntimeError("no default PipeWire source; set AUDIO_SOURCE")


def _whisper() -> tuple[str, str]:
    cli = os.getenv("WHISPER_CLI", "").strip()
    local = Path.home() / ".local/bin/whisper-cli"
    model = os.getenv(
        "WHISPER_MODEL_PATH",
        str(Path.home() / ".local/share/models/whisper/ggml-base.bin"),
    )
    return cli or (str(local) if local.is_file() else "whisper-cli"), model


def transcribe(path: str) -> str:
    cli, model = _whisper()
    proc = subprocess.run(
        [cli, "-m", model, "-f", path, "-l", "en", "--no-timestamps", "-np"],
        capture_output=True, text=True, timeout=90, check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError((proc.stderr or proc.stdout or "whisper-cli failed").strip())
    return " ".join(ln.strip() for ln in proc.stdout.splitlines() if ln.strip() and ln.strip() not in _BLANK)


class PushToTalk:
    def __init__(self, loop, queue, on_status) -> None:
        self.loop, self.queue, self.on_status = loop, queue, on_status
        self._stop, self._proc, self._path = threading.Event(), None, None
        self._ctrl = self._other = False
        self._started, self._target, self._mic_name = 0.0, "", ""

    def start(self) -> str | None:
        try:
            self._target, self._mic_name = mic_target()
            self._devices = _keyboards()
        except (OSError, subprocess.CalledProcessError, RuntimeError) as exc:
            return str(exc)
        if not self._devices:
            return "cannot read /dev/input. sudo usermod -aG input $USER, then log out."
        threading.Thread(target=self._loop, name="push-to-talk", daemon=True).start()
        return None

    def stop(self) -> None:
        self._stop.set()
        self._drop()

    def _loop(self) -> None:
        while not self._stop.is_set():
            readable, _, _ = select.select(self._devices, [], [], 0.2)
            for dev in readable:
                try:
                    events = dev.read()
                except OSError:
                    continue
                for event in events:
                    if event.type == EV_KEY:
                        self._on_key(event.code, event.value)

    def _on_key(self, code: int, value: int) -> None:
        if code == KEY_LEFTCTRL and value == 1 and not self._ctrl:
            self._ctrl, self._other = True, False
            self._begin()
        elif code == KEY_LEFTCTRL and value == 0 and self._ctrl:
            self._ctrl = False
            self._end()
        elif self._ctrl and value == 1:
            self._other = True
            self._drop()

    def _begin(self) -> None:
        self._drop()
        fd, self._path = tempfile.mkstemp(prefix="shanks-voice-", suffix=".wav")
        os.close(fd)
        self._started = time.monotonic()
        try:
            self._proc = subprocess.Popen(
                ["pw-record", f"--target={self._target}", "--rate=16000", "--channels=1", "--format=s16", self._path],
                stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
            )
        except FileNotFoundError:
            self.on_status("pw-record not found")
            return
        self.on_status(f"listening ({self._mic_name})")

    def _end(self) -> None:
        path, kept = self._take(), not self._other and time.monotonic() - self._started >= MIN_HOLD_S
        if path is None:
            return
        try:
            if not kept:
                return
            text = transcribe(path)
            if text:
                self.loop.call_soon_threadsafe(self.queue.put_nowait, text)
            else:
                self.on_status("heard nothing")
        except (OSError, subprocess.TimeoutExpired, RuntimeError) as exc:
            self.on_status(f"voice: {exc}")
        finally:
            Path(path).unlink(missing_ok=True)

    def _drop(self) -> None:
        path = self._take()
        if path:
            Path(path).unlink(missing_ok=True)

    def _take(self) -> str | None:
        path, self._path = self._path, None
        proc, self._proc = self._proc, None
        if proc and proc.poll() is None:
            proc.send_signal(signal.SIGINT)
            try:
                proc.wait(timeout=2)
            except subprocess.TimeoutExpired:
                proc.kill()
        return path


def _keyboards():
    from evdev import InputDevice, ecodes, list_devices

    found = []
    for path in list_devices():
        try:
            dev = InputDevice(path)
        except OSError:
            continue
        if ecodes.KEY_LEFTCTRL in dev.capabilities().get(ecodes.EV_KEY, []):
            found.append(dev)
    return found
