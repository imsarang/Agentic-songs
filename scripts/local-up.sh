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

start spotify env MCP_TRANSPORT=http MCP_ADDR=:8080 SPOTIFY_DATA_DIR="$DATA" "$BIN/spotify"
start google env MCP_TRANSPORT=http MCP_ADDR=:8004 GOOGLE_DATA_DIR="$DATA" "$BIN/google"

if ss -ltn | grep -q '127.0.0.1:8092'; then
  echo "machine already up"
else
  start machine "$BIN/machine"
fi

aud_py=python3
if [[ -x audio-analyst/.venv/bin/python ]]; then
  aud_py=$ROOT/audio-analyst/.venv/bin/python
fi
(
  cd "$ROOT/audio-analyst"
  start analyst env MCP_TRANSPORT=http MCP_ADDR=:8090 REDIS_URL=redis://127.0.0.1:6380/0 \
    "$aud_py" -m app.mcp_server
)

if [[ -x agent/.venv/bin/shanks-agent ]]; then
  start agent "$ROOT/agent/.venv/bin/shanks-agent"
else
  start agent env PYTHONPATH="$ROOT/agent/src" "$ROOT/agent/.venv/bin/python" -m shanks_agent
fi
