# One-word base, then -up or -down. Foreground commands have no suffix.
ROOT  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
UNIT  := host-capturer.service
SPOTIFY_BIN ?= /snap/bin/spotify
export SHANKS_DIR ?= $(HOME)/.config/shanks
DASH_URL := http://127.0.0.1:3000/d/shanks-live/shanks-live?orgId=1&from=now-15m&to=now&timezone=browser&refresh=5s

LOCAL_BIN := $(ROOT).local/bin
LOCAL_LOG := $(ROOT).local/logs
LOCAL_PID := $(ROOT).local/pids

.PHONY: help bins \
	stack-up stack-down stack-boot \
	host-up host-down redis-up \
	spotify-up spotify-down google-up google-down machine-up machine-down \
	retrieve-up retrieve-down mcp-up mcp-down agent-up agent-down ui \
	crawl-up crawl-down \
	embed-up embed-down worker-up worker-down \
	ingest-up ingest-down eval

help:
	@echo "make stack-up       qdrant, minio, crawl redis :6391, embed :8091, worker, agent :8001, retriever :8093, grafana, capturer, Chrome"
	@echo "make stack-down     stop that stack"
	@echo "make stack-boot     rebuild compose images, then stack-up"
	@echo "make host-up        all MCP servers, agent :8001, grafana, Chrome"
	@echo "make host-down      stop host-up"
	@echo "make spotify-up     Spotify MCP :8080"
	@echo "make spotify-down   stop Spotify MCP"
	@echo "make google-up      Google MCP :8004"
	@echo "make google-down    stop Google MCP"
	@echo "make machine-up     machine MCP :8092"
	@echo "make machine-down   stop machine MCP"
	@echo "make retrieve-up    retrieve MCP :8093"
	@echo "make retrieve-down  stop retrieve MCP"
	@echo "make mcp-up         spotify, google, machine, and retrieve"
	@echo "make mcp-down       stop every MCP server"
	@echo "make agent-up       agent :8001"
	@echo "make agent-down     stop the agent"
	@echo "make ui             terminal UI against the agent"
	@echo "make crawl-up       qdrant, minio :9000, crawl redis :6391, crawl"
	@echo "make crawl-down     stop the crawl stack"
	@echo "make embed-up       embed server :8091 in the background"
	@echo "make embed-down     stop the embed server"
	@echo "make worker-up      worker in the background"
	@echo "make worker-down    stop the worker"
	@echo "make ingest-up      crawl-up plus embed-up and worker-up"
	@echo "make ingest-down    stop ingest-up"
	@echo "make eval           run the RAGAS eval. Needs LM Studio, ingest-up, retrieve-up, and agent-up"
	@echo "make help           this list"

bins:
	mkdir -p $(LOCAL_BIN)

stack-up:
	mkdir -p $(HOME)/.config/systemd/user $(LOCAL_LOG) $(LOCAL_PID)
	cp $(ROOT)host/$(UNIT) $(HOME)/.config/systemd/user/$(UNIT)
	systemctl --user daemon-reload
	docker compose -f $(ROOT)docker-compose.crawl.yml up -d qdrant minio redis
	$(MAKE) embed-up worker-up
	docker compose -f $(ROOT)docker-compose.yml up -d --build --wait
	systemctl --user start $(UNIT)
	google-chrome --user-data-dir="$(HOME)/.config/shanks/grafana-chrome" --app='$(DASH_URL)' --no-first-run --no-default-browser-check >/dev/null 2>&1 &
	@echo "agent http://127.0.0.1:8001  retriever :8093  qdrant :6333  embed :8091  worker log $(LOCAL_LOG)/worker.log"

stack-down:
	-pkill -f 'shanks/grafana-chrom[e]'
	-systemctl --user stop $(UNIT)
	$(MAKE) embed-down worker-down
	-docker compose -f $(ROOT)docker-compose.yml exec -T redis redis-cli FLUSHALL*
	docker compose -f $(ROOT)docker-compose.yml down
	-docker compose -f $(ROOT)docker-compose.crawl.yml stop qdrant minio redis

stack-boot:
	docker compose -f $(ROOT)docker-compose.yml build
	$(MAKE) stack-up

redis-up:
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	@redis-cli -p 6380 ping >/dev/null 2>&1 || \
		redis-server --port 6380 --bind 127.0.0.1 --daemonize yes \
			--loadmodule /usr/lib/redis/modules/redistimeseries.so \
			--dir $(ROOT).local --dbfilename redis.rdb --pidfile $(LOCAL_PID)/redis.pid
	@redis-cli -p 6380 MODULE LIST | grep -q timeseries || \
		redis-cli -p 6380 MODULE LOAD /usr/lib/redis/modules/redistimeseries.so

spotify-up: bins
	mkdir -p -m 700 $(SHANKS_DIR) $(LOCAL_LOG) $(LOCAL_PID)
	@test -f $(SHANKS_DIR)/spotify-token.json || \
		{ test -f $(ROOT)mcp/data/spotify-token.json && cp $(ROOT)mcp/data/spotify-token.json $(SHANKS_DIR)/spotify-token.json && chmod 600 $(SHANKS_DIR)/spotify-token.json; true; }
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/spotify ./cmd/spotify
	@bash -eu $(ROOT)scripts/local-up.sh $(ROOT) $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID) $(SHANKS_DIR) spotify
	@echo "spotify MCP :8080"

spotify-down:
	-kill $$(cat $(LOCAL_PID)/spotify.pid) 2>/dev/null
	-rm -f $(LOCAL_PID)/spotify.pid

google-up: bins
	mkdir -p -m 700 $(SHANKS_DIR) $(LOCAL_LOG) $(LOCAL_PID)
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/google ./cmd/google
	@bash -eu $(ROOT)scripts/local-up.sh $(ROOT) $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID) $(SHANKS_DIR) google
	@echo "google MCP :8004"

google-down:
	-kill $$(cat $(LOCAL_PID)/google.pid) 2>/dev/null
	-rm -f $(LOCAL_PID)/google.pid

machine-up: bins
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	cd $(ROOT)mcp && go build -o $(LOCAL_BIN)/machine ./cmd/machine
	@bash -eu $(ROOT)scripts/local-up.sh $(ROOT) $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID) $(SHANKS_DIR) machine
	@echo "machine MCP :8092"

machine-down:
	-kill $$(cat $(LOCAL_PID)/machine.pid) 2>/dev/null
	-rm -f $(LOCAL_PID)/machine.pid

retrieve-up: redis-up
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	@bash -eu $(ROOT)scripts/local-up.sh $(ROOT) $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID) $(SHANKS_DIR) retriever
	@echo "retrieve MCP :8093"

retrieve-down:
	-kill $$(cat $(LOCAL_PID)/retriever.pid) 2>/dev/null
	-rm -f $(LOCAL_PID)/retriever.pid

mcp-up: spotify-up google-up machine-up retrieve-up

mcp-down: spotify-down google-down machine-down retrieve-down

agent-up: redis-up
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	@bash -eu $(ROOT)scripts/local-up.sh $(ROOT) $(LOCAL_BIN) $(LOCAL_LOG) $(LOCAL_PID) $(SHANKS_DIR) agent
	@echo "agent http://127.0.0.1:8001"

agent-down:
	-kill $$(cat $(LOCAL_PID)/agent.pid) 2>/dev/null
	-rm -f $(LOCAL_PID)/agent.pid

host-up: mcp-up agent-up
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	@bash -eu $(ROOT)scripts/local-grafana.sh $(ROOT)
	@echo "grafana http://127.0.0.1:3000  ui: make ui"
	google-chrome --user-data-dir="$(HOME)/.config/shanks/grafana-chrome" --app='$(DASH_URL)' --no-first-run --no-default-browser-check >/dev/null 2>&1 &

host-down:
	@bash -eu $(ROOT)scripts/local-down.sh $(LOCAL_PID)

ui:
	cd $(ROOT)ui && .venv/bin/shanks-ui

crawl-up:
	bash $(ROOT)scripts/crawl-up.sh

crawl-down:
	docker compose -f $(ROOT)docker-compose.crawl.yml down

embed-up:
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	@if [ ! -f $(LOCAL_PID)/embed.pid ] || ! kill -0 `cat $(LOCAL_PID)/embed.pid` 2>/dev/null; then \
		nohup bash $(ROOT)scripts/embed.sh >>$(LOCAL_LOG)/embed.log 2>&1 & echo $$! > $(LOCAL_PID)/embed.pid; \
	fi
	@echo "embed :8091  log $(LOCAL_LOG)/embed.log"

embed-down:
	-kill $$(cat $(LOCAL_PID)/embed.pid) 2>/dev/null
	-rm -f $(LOCAL_PID)/embed.pid

worker-up:
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	@if [ ! -f $(LOCAL_PID)/worker.pid ] || ! kill -0 `cat $(LOCAL_PID)/worker.pid` 2>/dev/null; then \
		nohup bash $(ROOT)scripts/worker.sh >>$(LOCAL_LOG)/worker.log 2>&1 & echo $$! > $(LOCAL_PID)/worker.pid; \
	fi
	@echo "worker log $(LOCAL_LOG)/worker.log"

worker-down:
	-kill $$(cat $(LOCAL_PID)/worker.pid) 2>/dev/null
	-rm -f $(LOCAL_PID)/worker.pid

ingest-up:
	mkdir -p $(LOCAL_LOG) $(LOCAL_PID)
	docker compose -f $(ROOT)docker-compose.crawl.yml up -d --build qdrant redis minio crawl
	$(MAKE) embed-up worker-up
	@echo "qdrant :6333  redis :6391  minio :9000  embed :8091  crawl + worker started"

ingest-down:
	$(MAKE) embed-down worker-down
	$(MAKE) crawl-down

eval:
	$(ROOT)evals/.venv/bin/python $(ROOT)evals/interview/src/main.py
