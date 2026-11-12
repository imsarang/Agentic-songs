#!/usr/bin/env bash
# Start the host processes on the ports compose publishes.
set -eu
ROOT=$1
BIN=$2
LOG=$3
PID=$4
DATA=$5
cd "$ROOT"

start() {
  name=$1
  shift
  if [[ -f $PID/$name.pid ]] && kill -0 "$(cat "$PID/$name.pid")" 2>/dev/null; then
    echo "$name already up"
    return
  fi
  nohup "$@" >>"$LOG/$name.log" 2>&1 &
  echo $! >"$PID/$name.pid"
  echo "$name pid $(cat "$PID/$name.pid")"
}

want=${6:-all}
aud_py=python3
if [[ -x mcp-py/.venv/bin/python ]]; then
  aud_py=$ROOT/mcp-py/.venv/bin/python
fi
if [[ $want == all || $want == spotify ]]; then
  start spotify env MCP_TRANSPORT=http MCP_ADDR=:8080 SPOTIFY_DATA_DIR="$DATA" "$BIN/spotify"
fi
if [[ $want == all || $want == google ]]; then
  start google env MCP_TRANSPORT=http MCP_ADDR=:8004 GOOGLE_DATA_DIR="$DATA" "$BIN/google"
fi
if [[ $want == all || $want == machine ]]; then
  start machine env MCP_TRANSPORT=http MCP_ADDR=:8092 "$BIN/machine"
fi
if [[ $want == all || $want == retriever ]]; then
  start retriever env MCP_TRANSPORT=http MCP_ADDR=:8093 \
    QDRANT_URL="${QDRANT_URL:-http://127.0.0.1:6333}" \
    EMBED_URL="${EMBED_URL:-http://127.0.0.1:8091/embed}" \
    PYTHONPATH="$ROOT/mcp-py" \
    "$aud_py" "$ROOT/mcp-py/cmd/retriever/main.py"
fi
if [[ $want == all || $want == agent ]]; then
  if [[ -x agent/.venv/bin/shanks-agent ]]; then
    start agent "$ROOT/agent/.venv/bin/shanks-agent"
  else
    start agent env PYTHONPATH="$ROOT/agent/src" "$ROOT/agent/.venv/bin/python" -m shanks_agent
  fi
fi
