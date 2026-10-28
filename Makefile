# make up / make down — no boot enable for the host ear
ROOT  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
UNIT  := host-capturer.service
SPOTIFY_BIN ?= /snap/bin/spotify
export SHANKS_SPOTIFY_DIR ?= $(HOME)/.config/shanks
DASH_URL := http://127.0.0.1:3000/d/shanks-live/shanks-live?orgId=1&from=now-15m&to=now&timezone=browser&refresh=5s

LOCAL_BIN := $(ROOT).local/bin
LOCAL_LOG := $(ROOT).local/logs
LOCAL_PID := $(ROOT).local/pids

.PHONY: up down boot local local-down local-bins local-ui minio local-crawl-down

up:
	mkdir -p -m 700 $(SHANKS_SPOTIFY_DIR)
	@test -f $(SHANKS_SPOTIFY_DIR)/spotify-token.json || \
		{ test -f $(ROOT)mcp/data/spotify-token.json && cp $(ROOT)mcp/data/spotify-token.json $(SHANKS_SPOTIFY_DIR)/spotify-token.json && chmod 600 $(SHANKS_SPOTIFY_DIR)/spotify-token.json; true; }
	mkdir -p $(HOME)/.config/systemd/user
	cp $(ROOT)host/$(UNIT) $(HOME)/.config/systemd/user/$(UNIT)
	systemctl --user daemon-reload
	pgrep -x spotify >/dev/null || nohup $(SPOTIFY_BIN) >/dev/null 2>&1 &
	docker compose -f $(ROOT)docker-compose.yml up -d --wait
	systemctl --user start $(UNIT)
	google-chrome --user-data-dir="$(HOME)/.config/shanks/grafana-chrome" --app='$(DASH_URL)' --no-first-run --no-default-browser-check >/dev/null 2>&1 &

down:
	-pkill -f 'shanks/grafana-chrom[e]'
	-pkill -x spotify
	-systemctl --user stop $(UNIT)
	-docker compose -f $(ROOT)docker-compose.yml exec -T redis redis-cli FLUSHALL*
	docker compose -f $(ROOT)docker-compose.yml down

boot:
	docker compose -f $(ROOT)docker-compose.yml build
	$(MAKE) up

# Host processes on the same ports compose publishes. No image build.
# Google listens on :8004 here; in compose that port is a map onto container :8080.
local: local-bins
	mkdir -p -m 700 $(SHANKS_SPOTIFY_DIR) $(LOCAL_LOG) $(LOCAL_PID)
	@test -f $(SHANKS_SPOTIFY_DIR)/spotify-token.json || \
		{ test -f $(ROOT)mcp/data/spotify-token.json && cp $(ROOT)mcp/data/spotify-token.json $(SHANKS_SPOTIFY_DIR)/spotify-token.json && chmod 600 $(SHANKS_SPOTIFY_DIR)/spotify-token.json; true; }
	@pgrep -x spotify >/dev/null || nohup $(SPOTIFY_BIN) >/dev/null 2>&1 &
	@redis-cli -p 6380 ping >/dev/null 2>&1 || \
		redis-server --port 6380 --bind 127.0.0.1 --daemonize yes \
			--loadmodule /usr/lib/redis/modules/redistimeseries.so \
			--dir $(ROOT).local --dbfilename redis.rdb --pidfile $(LOCAL_PID)/redis.pid
	@redis-cli -p 6380 MODULE LIST | grep -q timeseries || \
		redis-cli -p 6380 MODULE LOAD /usr/lib/redis/modules/redistimeseries.so
	@bash -eu $(ROOT)scripts/local-up.sh $(ROOT) $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID) $(SHANKS_SPOTIFY_DIR)
	@bash -eu $(ROOT)scripts/local-grafana.sh $(ROOT)
	@echo "agent http://127.0.0.1:8001  spotify :8080  google :8004  machine :8092  analyst :8090  redis :6380"
	@echo "grafana http://127.0.0.1:3000  ui: make local-ui"
	google-chrome --user-data-dir="$(HOME)/.config/shanks/grafana-chrome" --app='$(DASH_URL)' --no-first-run --no-default-browser-check >/dev/null 2>&1 &

local-ui:
	cd $(ROOT)ui && .venv/bin/shanks-ui

local-down:
	@bash -eu $(ROOT)scripts/local-down.sh $(LOCAL_PID)

minio:
	docker compose -f $(ROOT)docker-compose.crawl.yml up -d
	@echo "minio http://127.0.0.1:9000  console http://127.0.0.1:9001  crawl container shanks-crawl"

local-crawl-down:
	docker compose -f $(ROOT)docker-compose.crawl.yml down

local-bins:
	mkdir -p $(LOCAL_BIN)
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/spotify ./cmd/spotify
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/google ./cmd/google
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/machine ./cmd/machine