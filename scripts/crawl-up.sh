#!/usr/bin/env bash
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
docker compose -f "$ROOT/docker-compose.crawl.yml" up -d --build --remove-orphans
echo "minio http://127.0.0.1:9000  console http://127.0.0.1:9001  qdrant :6333  queue redis :6391"
