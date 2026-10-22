#!/usr/bin/env bash
cd "$(dirname "$0")"
export PYTHONPATH=. JACK_NO_START_SERVER=1
export REDIS_URL="${REDIS_URL:-redis://127.0.0.1:6380/0}"
exec .venv/bin/python -m capturer.publish
