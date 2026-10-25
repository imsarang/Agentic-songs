# make up / make down — no boot enable for the host ear
ROOT  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
UNIT  := host-capturer.service
SPOTIFY_BIN ?= /snap/bin/spotify
export SHANKS_SPOTIFY_DIR ?= $(HOME)/.config/shanks
DASH_URL := http://127.0.0.1:3000/d/shanks-live/shanks-live?orgId=1&from=now-15m&to=now&timezone=browser&refresh=5s

.PHONY: up down

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
