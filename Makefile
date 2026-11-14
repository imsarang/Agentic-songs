ROOT  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SPOTIFY_BIN ?= /snap/bin/spotify
export SHANKS_DIR ?= $(HOME)/.config/shanks
LOCAL_BIN := $(ROOT).local/bin
LOCAL_LOG := $(ROOT).local/logs
LOCAL_PID := $(ROOT).local/pids
UP := bash -eu $(ROOT)scripts/local-up.sh $(ROOT) $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID) $(SHANKS_DIR)

.PHONY: help boot down shanks ingest spotify google machine ui dirs redis

help:
	@echo "make boot     MCP servers, agent, crawl, worker, and the ear"
	@echo "make down     stop boot"
	@echo "make shanks   MCP servers and the agent"
	@echo "make ingest   crawl and the worker"
	@echo "make spotify  Spotify MCP :8080"
	@echo "make google   Google MCP :8004"
	@echo "make machine  machine MCP :8092"
	@echo "make ui       terminal UI"

boot: ingest shanks
	mkdir -p $(HOME)/.config/systemd/user
	cp $(ROOT)host/host-capturer.service $(HOME)/.config/systemd/user/host-capturer.service
	systemctl --user daemon-reload
	systemctl --user start host-capturer.service

down:
	-systemctl --user stop host-capturer.service
	-pkill -f '/snap/spotify/'
	@bash -eu $(ROOT)scripts/local-down.sh $(LOCAL_PID)
	-docker compose -f $(ROOT)docker-compose.yml down
	-docker compose -f $(ROOT)docker-compose.crawl.yml down

dirs:
	mkdir -p -m 700 $(SHANKS_DIR)
	mkdir -p $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID)

redis: dirs
	@redis-cli -p 6380 ping >/dev/null 2>&1 || \
		redis-server --port 6380 --bind 127.0.0.1 --daemonize yes \
			--loadmodule /usr/lib/redis/modules/redistimeseries.so \
			--dir $(ROOT).local --dbfilename redis.rdb --pidfile $(LOCAL_PID)/redis.pid
	@redis-cli -p 6380 MODULE LIST | grep -q timeseries || \
		redis-cli -p 6380 MODULE LOAD /usr/lib/redis/modules/redistimeseries.so

spotify: dirs
	@test -f $(SHANKS_DIR)/spotify-token.json || \
		{ test -f $(ROOT)mcp/data/spotify-token.json && cp $(ROOT)mcp/data/spotify-token.json $(SHANKS_DIR)/spotify-token.json && chmod 600 $(SHANKS_DIR)/spotify-token.json; true; }
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/spotify ./cmd/spotify
	@$(UP) spotify
	@pgrep -f '/snap/spotify/' >/dev/null || nohup $(SPOTIFY_BIN) >/dev/null 2>&1 &
	@echo "spotify MCP :8080"

google: dirs
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/google ./cmd/google
	@$(UP) google
	@echo "google MCP :8004"

machine: dirs
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/machine ./cmd/machine
	@$(UP) machine
	@echo "machine MCP :8092"

shanks: spotify google machine redis
	@$(UP) retriever
	@$(UP) agent
	@echo "agent http://127.0.0.1:8001  retrieve :8093"

ui:
	cd $(ROOT)ui && .venv/bin/shanks-ui

ingest: dirs
	docker compose -f $(ROOT)docker-compose.crawl.yml up -d --build qdrant redis minio crawl
	@if [ ! -f $(LOCAL_PID)/embed.pid ] || ! kill -0 `cat $(LOCAL_PID)/embed.pid` 2>/dev/null; then \
		nohup bash $(ROOT)scripts/embed.sh >>$(LOCAL_LOG)/embed.log 2>&1 & echo $$! > $(LOCAL_PID)/embed.pid; \
	fi
	@if [ ! -f $(LOCAL_PID)/worker.pid ] || ! kill -0 `cat $(LOCAL_PID)/worker.pid` 2>/dev/null; then \
		nohup bash $(ROOT)scripts/worker.sh >>$(LOCAL_LOG)/worker.log 2>&1 & echo $$! > $(LOCAL_PID)/worker.pid; \
	fi
	@echo "qdrant :6333  redis :6391  minio :9000  embed :8091  crawl + worker"
