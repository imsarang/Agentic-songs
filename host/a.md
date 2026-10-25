# Host folder — what it is

The `host/` directory is the **machine-local ear**. It listens to what your desktop is playing (PipeWire sink), turns that audio into a few numbers, and writes those numbers to Redis. The containerized `audio-analyst` only **reads** Redis; it never touches your sound system.

---

## Why this folder exists

- Docker / containers usually cannot see your PipeWire session cleanly.
- So capture stays on the host; analysis / MCP stays in the analyst service.
- Nothing here serves HTTP. It is a process that writes Redis keys.

---

## Layout

| Path | Role |
|------|------|
| `sample.sh` | One capture cycle: run Python once, then exit |
| `run.sh` | Forever loop that keeps calling `sample.sh` |
| `capturer/capture.py` | Record a short PCM window from the default sink (`pw-record`) |
| `capturer/features.py` | PCM → `loudness_db`, `energy`, `silence` (no waveform kept) |
| `capturer/publish.py` | Glue: capture → features → Redis |
| `requirements.txt` | `numpy`, `redis` |
| `.venv/` | Local Python env (create once; not committed as source of truth) |
| `host-capturer.service` | Optional systemd unit that starts `run.sh` |

---

## The shell scripts (keep them thin)

### `sample.sh` — one wire sample

1. `cd` into `host/`.
2. Set `PYTHONPATH=.` so `python -m capturer.publish` works.
3. Set `JACK_NO_START_SERVER=1` so PipeWire/JACK helpers do not spawn a stray JACK server.
4. Point `REDIS_URL` at Redis (default `redis://127.0.0.1:6380/0`).
5. `exec` `.venv/bin/python -m capturer.publish`.

Shell does **not** do math or audio. It only launches Python.

### `run.sh` — keep sampling

1. `cd` into `host/`.
2. Loop forever: run `./sample.sh`.
3. If a sample fails, sleep 2 seconds and try again.

That is the long-running “daemon.” No extra logic.

---

## What Python does each cycle

1. **Capture** (`capture.py`)
   - Find the default PipeWire **sink** (`wpctl status`, or override with `AUDIO_TARGET`).
   - Run `pw-record` for about `WINDOW_MS` (default 5000 ms) at `SAMPLE_RATE` (default 44100).
   - Return float32 PCM in memory (not saved to disk).

2. **Features** (`features.py`)
   - RMS → `loudness_db` (`20 * log10(rms)`).
   - Scaled RMS → `energy` (0–1-ish).
   - `silence` if loudness &lt; −50 dB.
   - Waveform is discarded after this.

3. **Publish** (`publish.py`)
   - `SET analyst:wire:latest` → JSON `{ ts, features, device }` (latest snapshot for the analyst).
   - `TS.ADD live:<metric>` → RedisTimeSeries for Grafana (`series.py`).

---

## Redis contract (what the analyst expects)

- **Key** `analyst:wire:latest` (string JSON) — newest feature blob. This is what the analyst reads.
- **Time series** `live:loudness_db`, `live:peak_db`, `live:bass`, `live:mid`, `live:high`, `live:energy`, `live:brightness`, `live:crest`, `live:beat_hits`, `live:vibe`, `live:tempo_bpm`, `live:mood_code` — Grafana.

Redis itself does **not** store PCM or waveforms—only the feature numbers.

---

## One-time setup

```bash
cd host
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
```

Host tools needed: `pw-record`, `wpctl` (PipeWire), and Redis reachable on the URL you set (compose often exposes `6380`).

---

## How to run

- **Stack + ear (recommended):** from repo root, `make up` / `make down`  
  - starts Docker Compose, then `systemctl --user start` the host wire  
  - does **not** enable the unit on boot — ear only runs while you `make up`
- **One sample:** `./sample.sh`
- **Continuous ear alone:** `./run.sh` or `systemctl --user start host-capturer.service` (unit lives in `~/.config/systemd/user/`)

---

## Useful env vars

| Var | Default | Meaning |
|-----|---------|---------|
| `REDIS_URL` | `redis://127.0.0.1:6380/0` | Where to write |
| `WIRE_KEY` | `analyst:wire:latest` | Latest JSON key |
| `HITS_KEY` | `analyst:live:hits` | Recent beat times for tempo |
| `WINDOW_MS` | `1000` | Capture window length |
| `SAMPLE_RATE` | `44100` | `pw-record` rate |
| `AUDIO_TARGET` | (auto via `wpctl`) | Force a sink id/name |

---

## Mental model

```
desktop audio (PipeWire sink)
        ↓  pw-record (host)
   PCM in memory
        ↓  features.py
   loudness / energy / silence
        ↓  Redis
   audio-analyst container (MCP / decisions)
```

Shell = start Python. Python = hear + measure + publish. Analyst = consume Redis.
