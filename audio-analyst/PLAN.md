# Audio Analyst Service — Plan

A small **Python** sidecar that listens to real audio (what’s playing on the machine), derives features, and exposes decisions/signals the Spotify MCP (or a future client) can use.

This replaces Spotify `/audio-features` (blocked on Development Mode) with **our own** analysis of the live signal.

---

## Goal

| Need | Who owns it |
|------|-------------|
| Play / pause / playlists / OAuth | Existing `spotify` MCP (Go) |
| “What does this sound like?” (energy, tempo-ish, vibe) | New `audio-analyst` (Python) |
| Act on that (skip, loudness warn, mood tag, …) | Analyst decides → MCP/client consumes |

---

## High-level shape

```
┌─────────────────┐     HTTP/MCP      ┌──────────────────┐
│  spotify (Go)   │ ◄───────────────► │ client / Cursor  │
└────────┬────────┘                   └────────▲─────────┘
         │ play / now-playing                    │
         ▼                                       │ features / decisions
┌─────────────────┐     HTTP                     │
│ audio-analyst   │ ─────────────────────────────┘
│ (Python)        │
└───────┬───┬─────┘
        │   │
        │   ├── Redis (now): hot cache / in-progress samples
        │   └── DB (later): durable results for known tracks
        │ capture
        ▼
  host audio (PipeWire / Pulse / ALSA)
```

Analyst does **not** replace Spotify control. It only **hears** and **reports**.

**Lookup before listen:** unique **track id/URI** → Redis → (later) DB → only if miss, run the 5-sample capture + weighted avg, then write back.

**Identity rule:** never key by song **title** alone. The same name can be live vs studio, remix, cover, different masters — different sound. Always use Spotify’s unique track id, e.g. `spotify:track:5lWSa1rmuSL6OBPOnkAqoa` (from now-playing / play result). Optional display fields (`title`, `artist`) are metadata only, not cache keys.

---

## Repo layout (proposed)

```
audio-analyst/
  PLAN.md                 ← this file
  README.md               ← run instructions (later)
  Dockerfile
  requirements.txt
  app/
    main.py               ← HTTP API entry
    capture.py            ← open audio device / loopback
    features.py           ← compute metrics from PCM
    decisions.py          ← rules → structured actions
    store.py              ← Redis now; DB adapter later
    models.py             ← request/response shapes
  docker-compose.override snippet → merge into root compose
```

Keep it flat. No frameworks beyond what’s needed (e.g. FastAPI + sounddevice/numpy + redis).

---

## Storage

| Layer | When | Role |
|-------|------|------|
| **Redis** | v1 (now) | Immediate results + in-progress sample counters; keyed by **Spotify track URI/id** |
| **DB** (Postgres/SQLite) | later | Durable store for known **track ids** so repeat plays of that recording skip capture |

**Redis keys (sketch):** `{id}` = full URI or bare Spotify track id (pick one form and normalize)
- `analyst:track:{id}:partial` — list/hash of samples 1…N while sampling  
- `analyst:track:{id}:final` — weighted-avg features + bullets + decisions (TTL optional, e.g. 30d)  
- `analyst:track:{id}:meta` — `sample_i`, `status=sampling|done`, plus optional `title`/`artist` for humans  

**Flow:**
1. Request **must** include unique `track_uri` (or id) from Spotify now-playing — not just a name  
2. If `final` in Redis (or DB later) for that id → return cached (fast path)  
3. Else capture next sample, update `partial`, bump counter  
4. On Nth sample → weighted avg → write `final` under that id → return bullets  
5. Later migration: on `final`, also upsert DB by track id; lookup Redis → DB → capture  

Same title, different track id ⇒ **separate** cache rows (different style/performance is expected).  
Do not store raw PCM in Redis/DB — features + decisions only.

---

## What it listens to

**Preferred (Linux desktop):** monitor/loopback of the default sink  
(PipeWire/Pulse “monitor of …” = what speakers/headphones hear).

**Fallback:** a named input device (mic) — worse for “what Spotify is playing.”

Document one supported path first (PipeWire on the host), then ALSA.

---

## Features to compute (v1)

Map roughly to the old Spotify ideas, but from PCM:

| Signal | How (simple) |
|--------|----------------|
| loudness | RMS / dBFS |
| energy | short-window RMS variance |
| tempo (rough) | onset / beat estimate (librosa or a11y beat tracker) |
| brightness | spectral centroid |
| speechiness-ish | high zero-crossing / flat spectrum heuristic |
| live-ish | sudden audience-like noise — optional later |

v1 can ship **loudness + energy + tempo estimate** only. Expand later without changing the API envelope.

---

## Decisions (v1)

Small, explicit rules — not a black-box model:

- `too_loud` → suggest lower volume / pause  
- `high_energy` → tag mood `hype`  
- `low_energy` → tag mood `chill`  
- `silence` → “nothing useful on the wire”

Return JSON the LLM or Go MCP can turn into bullets.

Example response:

```json
{
  "ok": true,
  "features": {
    "loudness_db": -18.2,
    "energy": 0.71,
    "tempo_bpm": 122
  },
  "decisions": [
    { "type": "mood", "value": "hype" },
    { "type": "suggest", "value": "keep_playing" }
  ],
  "summary_bullets": [
    "High energy (~0.71)",
    "Tempo around 122 BPM",
    "Comfortable loudness"
  ]
}
```

---

## HTTP API (minimal)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/healthz` | liveness |
| GET | `/features` | analyze last N ms of capture |
| GET | `/decisions` | features + rule output |
| POST | `/capture/start` | start background ring buffer |
| POST | `/capture/stop` | stop capture |

Auth: none on local docker network; bind to internal port only.

---

## Docker & host audio

Containers do **not** see host audio by default. Plan for Linux + PipeWire/Pulse:

1. Run analyst with access to the host sound server socket.  
2. Prefer **host network** or mount the native socket.  
3. Pass device name via env.

### Compose sketch (root `docker-compose.yml`)

```yaml
services:
  audio-analyst:
    build: ./audio-analyst
    environment:
      AUDIO_BACKEND: pulse          # pulse | pipewire | alsa
      AUDIO_SOURCE: default         # or explicit monitor name
      ANALYST_ADDR: ":8090"
    # Option A — simplest on a desktop host:
    network_mode: host
    volumes:
      - ${XDG_RUNTIME_DIR}/pulse:/run/user/1000/pulse:ro
      # PipeWire variant may mount /run/user/1000/pipewire-0
    group_add:
      - audio
    # Option B (later): non-host network + published 8090
    # ports: ["8090:8090"]

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - analyst-redis:/data

  spotify:
    # existing service…
    environment:
      AUDIO_ANALYST_URL: http://127.0.0.1:8090   # if host network
      # or http://audio-analyst:8090 if on the same bridge network

volumes:
  analyst-redis:
```

`audio-analyst` also gets `REDIS_URL: redis://redis:6379/0` (use `host.docker.internal` / `127.0.0.1` if analyst uses `network_mode: host`).

**Reality check:** reliable capture almost always means **host audio stack + socket mount** (or `network_mode: host`). Pure isolated containers without host device access will not hear Spotify.

Document a “dev on host, no Docker” run path too (`python -m app.main`) for faster iteration.

---

## Integration with Spotify MCP

Two nodes, one flow:

1. **Node A — Spotify MCP:** play / next / playlist play → returns `uri` + now-playing.  
2. **Node B — Audio analyst tool:** always invoked after a successful play-like step, with that `track_uri`.  
3. Agent shows play confirmation + analyst bullets (or “sampling 2/5…” until weighted final is ready).

Analyst stays a **separate tool** (direct HTTP or its own MCP entry). Do **not** fold it into the Go Spotify server.

Do **not** block play on analyst failure — if Node B errors, Node A’s play still counts.

---

## Phased delivery

### Phase 0 — Skeleton
- Folder + Dockerfile + FastAPI `/healthz`
- Compose service that starts (even without real audio)

### Phase 1 — Capture
- Read PCM from monitor device into a ring buffer
- One-shot slice read of `WINDOW_MS`

### Phase 2 — Multi-sample + weighted avg + Redis
- Per-track counter in Redis; store samples under `partial`
- `/decisions` returns `{status: sampling, i, n}` or cached/final weighted result
- On complete → write `final` to Redis
- Silence → weight 0

### Phase 3 — Decisions
- Rules in `decisions.py` on the **weighted** vector
- `summary_bullets` from the final result only

### Phase 4 — Tempo / richer features
- Add beat/tempo estimate
- Tune `SAMPLE_GAP_MS` / silence floor

### Phase 5 — MCP glue
- Env + optional call from Go after play/control
- Cursor/agent instructions: prefer analyst bullets over Spotify audio-features

### Phase 6 — DB persistence
- Postgres or SQLite table `track_analytics (track_id PK, title, artist, features_json, decisions_json, updated_at)` — **PK = Spotify track id/URI**, never title
- Lookup order: Redis → DB → capture
- On new `final`, write-through to DB (Redis remains hot cache)

---

## Env vars

| Var | Default | Meaning |
|-----|---------|---------|
| `ANALYST_ADDR` | `:8090` | listen address |
| `AUDIO_BACKEND` | `pulse` | pulse / pipewire / alsa |
| `AUDIO_SOURCE` | `default` | device or monitor name |
| `WINDOW_MS` | `5000` | length of **one** capture slice |
| `SAMPLE_COUNT` | `5` | how many slices per song before combining (then stop) |
| `SAMPLE_GAP_MS` | `15000` | wait between slices (spread past intro; tune later) |
| `SAMPLE_WEIGHTS` | `1,2,3,4,5` | weights for weighted avg (later = more; silence → 0) |
| `SAMPLE_RATE` | `44100` | capture rate |
| `REDIS_URL` | `redis://127.0.0.1:6379/0` | hot cache for partial + final |
| `FINAL_TTL_SEC` | `2592000` | optional Redis TTL for finals (~30d); 0 = no expiry |
| `DATABASE_URL` | _(later)_ | durable store when Phase 6 lands |

---

## Non-goals (for now)

- No browser / Playwright  
- No replacing Spotify OAuth or playback  
- No training ML models on Spotify content  
- No requirement for GPU  
- No macOS/Windows audio story until Linux path works  

---

## Success criteria

1. With Spotify playing on the host, `GET /decisions` returns non-silence features.  
2. Bullets are readable by an LLM/user without raw dumps.  
3. `docker compose up` starts `audio-analyst` beside `spotify`.  
4. If capture fails, service stays up and returns a clear error JSON (play still works).  

---

## FAQ — 20 questions (decisions + why)

### 1. Why build this at all?
Spotify `/audio-features` returns **403** on Development Mode. We still want energy/tempo/vibe-style signals. Hearing the **actual output** is the only path we control.

### 2. Why a separate Python service instead of putting this in the Go MCP?
Audio capture + DSP libraries (sounddevice, numpy, librosa) are far easier in Python. Keep Go for Spotify API/MCP; keep Python for PCM math. Clear boundary, fewer giant binaries.

### 3. Why not scrape open.spotify.com for features?
Fragile, against the “API not DOM” direction we already chose, and still wouldn’t give official audio-features. Live PCM is honest about what’s on the speakers.

### 4. Why containerize if audio needs the host?
So it starts with `docker compose`, shares the same project lifecycle as `spotify`, and has a repeatable image. Capture still uses **host** PipeWire/Pulse via a socket mount — the container is packaging, not a fake sound card.

### 5. PipeWire or Pulse on this machine?
**Both sockets exist:** `/run/user/1000/pipewire-0` and `/run/user/1000/pulse`. Default plan: **Pulse-compatible API** (`AUDIO_BACKEND=pulse`) talking to PipeWire’s Pulse shim — widest Python library support. Mount both if needed.

### 6. Exact mounts for this host?
- `XDG_RUNTIME_DIR=/run/user/1000`
- Pulse: `${XDG_RUNTIME_DIR}/pulse` → `/run/user/1000/pulse`
- PipeWire: `${XDG_RUNTIME_DIR}/pipewire-0` (and manager socket if required)
- UID **1000** (`strawhat`); `group_add: [audio]` when using ALSA devices under `/dev/snd`

### 7. Why `network_mode: host` in the sketch?
Simplest way for the analyst to reach the host sound server and for Cursor to hit `127.0.0.1:8090`. Bridge networking is Phase-later; host network is fine for a personal desktop stack.

### 8. Should Cursor call the analyst directly or only via Spotify MCP?
**Direct — decided.** Analyst is its **own tool / node** in the agent graph (HTTP now; optional MCP wrapper later). Whenever a song is played (`spotify_play` / control next / playlist play), the agent **also** invokes this node with the track URI. It stays a separate service so playback and analysis can evolve independently; Spotify MCP does not proxy analyst responses.

Optional later: expose the same API as an MCP tool (`audio_decisions`) still backed by this Python service — still not inside the Go Spotify binary.

### 9. Should decisions auto-call `spotify_control` (skip/pause)?
**No in v1.** Analyst only **suggests**. Why: auto-skip from a loudness heuristic will feel buggy and scary. Human/agent approves actions.

### 10. Mic vs sink monitor — which audio source?
**Sink monitor** (“what the speakers hear”). Why: that’s Spotify’s output. A mic hears the room, noise, and keyboard — bad for music features.

### 11. Why a ring buffer + window (e.g. 5s) instead of whole-song analysis?
We don’t have the file; we have a live stream. A short window gives near-real-time answers while the track plays. Whole-song needs download/decode we don’t have.

**Policy: N samples per song, then average (default N=5).**  
A single early window often catches silence/intro and misreads tempo/energy. So for each track:

1. Take **5 slices** (each ~`WINDOW_MS`, e.g. 5s) spaced through the listen (or until skip/next).  
2. Keep a **counter** `sample_i / SAMPLE_COUNT` and per-sample feature vectors.  
3. After the 5th sample (or track change / stop), combine with a **weighted average** → one final result + bullets.  
4. Do **not** keep sampling forever; cap at `SAMPLE_COUNT` per `track_uri`.

**Weights (default):** later slices count more (intro-safe), e.g. for 5 samples  
`w = [1, 2, 3, 4, 5]` → `final = Σ(wᵢ·fᵢ) / Σ(wᵢ)`.  
Near-silence samples get weight **0** (or are dropped) so empty intros don’t pull energy/tempo down.

```
play track
  sample 1/5 → features₁  (weight 1)
  sample 2/5 → features₂  (weight 2)
  …
  sample 5/5 → features₅  (weight 5)
  final = weighted_mean(features, weights)  → summary_bullets
```

### 12. Why these features first (loudness, energy, tempo)?
They map to the useful part of old Spotify analytics, are computable from PCM, and don’t need licensed ML models. Brightness/speechiness-ish can wait.

### 13. Why rules in `decisions.py` instead of an ML model?
Debuggable, no training on Spotify content (policy risk), works offline, enough for “hype / chill / too loud.” Add ML only if rules fail in practice.

### 14. Why HTTP JSON instead of making analyst an MCP server too?
One MCP (Spotify) is enough complexity. Analyst is a dumb internal API. Cursor can still call HTTP; later we can wrap it as MCP if we want.

### 15. Why not block `spotify_play` until analyst returns?
Playback must stay reliable. Analyst can be down, muted, or wrong device. Play succeeds; analytics are best-effort.

### 16. Why Docker *and* a host `python -m app.main` path?
Docker = compose lifecycle. Host Python = fast debug of devices/permissions without rebuilds. Same code, two run modes.

### 17. Will headless Chromium “player” audio feed the analyst?
Only if that container’s audio is routed into the **host** sink you monitor. Today’s player sidecar is for Connect device presence, often silent/headless. Prefer analyzing **host Spotify desktop/web** output. Don’t depend on `shanks-player-1` for PCM.

### 18. Privacy / policy — is this OK?
We analyze **local PCM the user is already playing**, for personal use. We do not scrape Spotify’s feature API or train models on Spotify catalogs. Still: don’t log raw audio to disk in v1.

### 19. What does success look like on day one?
Spotify playing on the host → `GET /decisions` → non-silence loudness/energy → a few bullets the agent can show. If capture fails → clear JSON error, Spotify MCP unaffected.

### 20. Recommended build order (why this sequence)?
1. `/healthz` in Docker (prove compose)  
2. Capture + loudness (prove audio path — hardest risk)  
3. Decisions/bullets (product value)  
4. Tempo (nice-to-have)  
5. MCP glue (convenience)  
Why: fail fast on device permissions before writing rules or Go integration.

---

## Linux architecture — files, buffers, components

What actually moves sound from Spotify to our weighted-average features on this stack (PipeWire + Pulse shim + Docker).

```
Spotify / browser
       |
       v
PipeWire daemon (user session graph)
       |
       v
Sink (speakers/HDMI) --monitor tap--> Pulse shim socket
                                            |
                                            v
                                   audio-analyst (Python)
                                   slice buffer x5 -> weighted avg
                                            |
                                            v
                                   Redis (+ DB later)
```

### Host session / sockets

| Path / thing | What it is | Why involved |
|--------------|------------|--------------|
| `/run/user/1000/` (`$XDG_RUNTIME_DIR`) | Per-user runtime dir for the logged-in session | PipeWire/Pulse sockets live here; containers must see **this user’s** audio, not root’s empty session |
| `/run/user/1000/pipewire-0` | PipeWire client socket | Native graph control/streams; optional direct PW clients |
| `/run/user/1000/pipewire-0-manager` | PipeWire manager socket | Some tools need it for device enumeration |
| `/run/user/1000/pulse/` | Pulse-compatible socket dir | Python libs often use Pulse API; PipeWire provides this shim — **preferred capture entry** |
| `pipewire` / `wireplumber` (systemd --user) | Daemons building the graph + policy | Without them, no sink/monitor exists for Spotify or us |

### Kernel / hardware

| Path / thing | What it is | Why involved |
|--------------|------------|--------------|
| `/dev/snd/*` (`controlC*`, `pcmC*D*p`, …) | ALSA device nodes | Final hardware I/O; PipeWire opens these, not our app (unless ALSA backend) |
| ALSA pcm buffers | Kernel ring buffers for DMA to the card | Absorbs timing jitter between userspace and hardware |
| `audio` group | GID allowed to open `/dev/snd` | Needed if we ever talk ALSA directly; Pulse/PW path usually goes via session socket instead |

### Graph objects (PipeWire concepts)

| Component | What it is | Why involved |
|-----------|------------|--------------|
| **Node** (Spotify stream) | One app playback stream in the graph | Source of the music we want to analyze |
| **Sink** (default output) | Speakers / headphones / HDMI | Where audio is mixed for listening |
| **Monitor** of the sink | Read-only tap of what’s going to the sink | Lets us hear “what’s playing” without a mic |
| **Links** | Connections between nodes | Route Spotify → sink; monitor ← sink |

### Inside audio-analyst

| Component | What it is | Why involved |
|-----------|------------|--------------|
| Capture client (sounddevice / Pulse) | Opens the monitor as an input stream | Pulls PCM into userspace |
| **Ring / slice buffer** (~`WINDOW_MS` of samples) | Short float/PCM array in process memory | Holds one analysis window; not whole song; not written to disk |
| Sample counter + partial list | State for slices 1…5 per `track_id` | Multi-sample policy before weighted avg |
| Feature compute (numpy/librosa-ish) | Loudness / energy / tempo from the slice | Replaces blocked Spotify `/audio-features` |
| Weighted average | `Σ wᵢfᵢ / Σ wᵢ` (silence w=0) | Intro-safe final vector per unique track id |

### Persistence / orchestration

| Component | What it is | Why involved |
|-----------|------------|--------------|
| **Redis** | Hot key-value store | Immediate `partial` / `final` by Spotify track id; skip re-listen when known |
| **DB** (later) | Durable table keyed by track id | Survives Redis flush; same fast path for known recordings |
| Docker volume mounts of `$XDG_RUNTIME_DIR/...` | Bind host sockets into the container | Container has no audio otherwise |
| `network_mode: host` (or published ports) | How Cursor/MCP reach `:8090` | Host audio + simple local HTTP |

### What we deliberately do **not** use for v1

| Thing | Why skip |
|-------|----------|
| Microphone device | Hears room noise, not clean Spotify PCM |
| Saving raw WAV/PCM to disk | Privacy + size; features JSON is enough |
| Spotify `/audio-features` | 403 on Dev Mode |
| Chromium `player` container as the only source | Often headless/silent; Connect device ≠ reliable monitor feed |

### End-to-end “why this path”

Spotify must play through the **host** PipeWire sink so a **monitor** exists. Analyst attaches there, copies **~5s slices** into a small buffer five times, averages with weights, stores under the **track URI** in Redis. Hardware ALSA stays under PipeWire; we only need the session sockets + in-memory buffers + Redis.
