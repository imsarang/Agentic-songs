# Shanks Agent + UI (design only)

Three services: terminal **UI**, Python **agent** (ReAct / LangGraph), Go **Spotify MCP**.

```
[ ui ]  --WebSocket-->  [ agent :8000 ]  --HTTP-->  [ spotify :8080/mcp ]
                              |
                         LLMs + LangSmith
```

## Who owns what

- **ui** — keyboard, layout, display. No LLM, no MCP, no LangGraph.
- **agent** — ReAct loop, LLM failover, MCP tool calls, LangSmith. No terminal drawing.
- **spotify** — Spotify tools only (existing Go server).

## Agent folder structure (empty placeholders)

```
agent/
  Dockerfile
  pyproject.toml
  .env.example
  src/shanks_agent/
    __main__.py
    config.py
    observability.py
    api/
      app.py
      schemas.py
    llm/
      protocol.py
      providers.py
      failover.py
    mcp/
      session.py
    graph/
      state.py
      build.py
      nodes/
        agent.py
        tools.py          # TOOL_NAMES registry → one node per MCP tool
```

**Expand later:** new MCP tool → `TOOL_NAMES` in `tools.py`. New LLM → `providers.py`.

## UI folder structure (empty placeholders)

```
ui/
  Dockerfile
  pyproject.toml
  src/shanks_ui/
    __main__.py
    config.py
    app.py
    transport.py
```

## WebSocket contract (`/ws`)

Client → server: `user_message` + `text` + `thread_id`  
Server → client: `tool_start` / `tool_end` / `assistant_message` / `error` / `turn_done`

## LangChain vs LangGraph vs LangSmith

- **LangChain** — models/messages inside the agent node.
- **LangGraph** — ReAct graph (agent ↔ per-tool nodes).
- **LangSmith** — traces from the agent service.

## Phases

### Phase 0 — Structure (done)
- Empty `agent/` + `ui/` folders
- This design doc
- No runnable code yet

### Phase 1 — Agent skeleton
- `pyproject.toml`, `config.py`, `Dockerfile`
- FastAPI app: `/health` only
- Run with uvicorn locally
- **Exit:** `GET /health` → `{"ok": true}`

### Phase 2 — WebSocket API
- `/ws` accepts `user_message`
- Echo / stub replies (`assistant_message` + `turn_done`)
- Schemas in `api/schemas.py`
- **Exit:** UI-less test with websockets client; one turn round-trips

### Phase 3 — MCP client
- `mcp/session.py`: connect to `http://localhost:8080/mcp` (or `spotify:8080` in Docker)
- `list_tools` + `call_tool`
- Optional `GET /mcp/tools` for smoke
- **Exit:** agent lists the six Spotify tools against running Go MCP

### Phase 4 — LLM + failover
- `llm/protocol.py`, `providers.py` (OpenAI, Anthropic), `failover.py`
- Config: `LLM_PROVIDERS=openai,anthropic`
- **Exit:** unit test / small script: primary fails → secondary answers

### Phase 5 — LangGraph ReAct
- `graph/state.py`, `nodes/agent.py`, `nodes/tools.py` (`TOOL_NAMES` → one node each)
- `build.py`: agent ↔ tool nodes ↔ agent until finish
- Wire `run_turn()` into `/ws` (emit `tool_start` / `tool_end` / `assistant_message`)
- **Exit:** “play X” from a WS client drives search + play via MCP

### Phase 6 — Terminal UI
- `ui/` Rich REPL + `transport.py`
- Talks only to `ws://agent:8000/ws`
- **Exit:** type in UI → see tool events + final answer

### Phase 7 — Docker Compose
- `agent` + `ui` services next to `spotify`
- `MCP_URL=http://spotify:8080/mcp`, `AGENT_WS_URL=ws://agent:8000/ws`
- **Exit:** `compose up spotify agent` + `compose run ui` works end-to-end

### Phase 8 — LangSmith + harden
- `observability.py`: tracing env, tags, `thread_id`
- Timeouts, clearer `error` events, WS reconnect in UI
- **Exit:** traces show `agent` + `tool_spotify_*` spans; bad MCP/LLM fails cleanly in UI

---

Code stays out of the repo until you ask for it to be written to disk.
