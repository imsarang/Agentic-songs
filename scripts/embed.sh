#!/usr/bin/env bash
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export EMBED_MODEL_PATH="${HOME}/.local/models/bge-large-en-v1.5"
exec worker/.venv/bin/python embed/serve.py
