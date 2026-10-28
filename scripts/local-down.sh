#!/usr/bin/env bash
set -eu
PID=$1
if redis-cli -p 6380 ping >/dev/null 2>&1; then
  redis-cli -p 6380 FLUSHALL
fi
if [[ -f $PID/redis.pid ]]; then
  redis-cli -p 6380 shutdown nosave || true
  rm -f "$PID/redis.pid" "$PID/../redis.rdb"
fi
docker stop shanks-grafana >/dev/null 2>&1 || true
for name in spotify google machine analyst agent; do
  file=$PID/$name.pid
  if [[ -f $file ]]; then
    kill "$(cat "$file")" 2>/dev/null || true
    rm -f "$file"
  fi
done
