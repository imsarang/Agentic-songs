# make up / make down — no boot enable for the host ear
ROOT  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
UNIT  := host-capturer.service
SPOTIFY_BIN ?= /snap/bin/spotify
export SHANKS_SPOTIFY_DIR ?= $(HOME)/.config/shanks

.PHONY: up down

up:
	mkdir -p -m 700 $(SHANKS_SPOTIFY_DIR)
	@test -f $(SHANKS_SPOTIFY_DIR)/spotify-token.json || \
		{ test -f $(ROOT)mcp/data/spotify-token.json && cp $(ROOT)mcp/data/spotify-token.json $(SHANKS_SPOTIFY_DIR)/spotify-token.json && chmod 600 $(SHANKS_SPOTIFY_DIR)/spotify-token.json; true; }
	mkdir -p $(HOME)/.config/systemd/user
	cp $(ROOT)host/$(UNIT) $(HOME)/.config/systemd/user/$(UNIT)
	systemctl --user daemon-reload
	pgrep -x spotify >/dev/null || nohup $(SPOTIFY_BIN) >/dev/null 2>&1 &
	docker compose -f $(ROOT)docker-compose.yml up -d
	systemctl --user start $(UNIT)

down:
	-pkill -x spotify
	-systemctl --user stop $(UNIT)
	-docker compose -f $(ROOT)docker-compose.yml exec -T redis redis-cli FLUSHALL
	docker compose -f $(ROOT)docker-compose.yml down
