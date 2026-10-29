#!/usr/bin/env bash
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export S3_ENDPOINT=http://127.0.0.1:9000
export S3_BUCKET=shanks-raw
export S3_ACCESS_KEY=shanks
export S3_SECRET_KEY=shanks-secret
export S3_REGION=us-east-1
export CELERY_BROKER_URL=redis://127.0.0.1:6391/1
export QDRANT_URL=http://127.0.0.1:6333
export EMBED_MODEL_PATH="${HOME}/.local/models/bge-large-en-v1.5"
export PYTHONPATH=worker/src
exec worker/.venv/bin/python -m worker
