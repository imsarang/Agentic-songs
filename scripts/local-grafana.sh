#!/usr/bin/env bash
# Grafana on the host network so 127.0.0.1:6380 is the make-local Redis.
set -eu
ROOT=$1
name=shanks-grafana
if docker ps --format '{{.Names}}' | grep -qx "$name"; then
  echo "grafana already up"
  exit 0
fi
if docker ps -a --format '{{.Names}}' | grep -qx "$name"; then
  docker start "$name" >/dev/null
  echo "grafana started"
  exit 0
fi
docker run -d --name "$name" --network host \
  -e GF_INSTALL_PLUGINS=redis-datasource \
  -e GF_AUTH_ANONYMOUS_ENABLED=true \
  -e GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer \
  -v "$ROOT/grafana/provisioning/local-datasources:/etc/grafana/provisioning/datasources:ro" \
  -v "$ROOT/grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards:ro" \
  grafana/grafana:11.4.0 >/dev/null
echo "grafana created"
