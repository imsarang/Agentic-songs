FROM python:3.12-slim

# No PipeWire — ear is ./host on the desktop (CAPTURE_URL).
WORKDIR /app
COPY requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --timeout 300 --retries 10 -r requirements.txt
COPY app ./app

ARG UID=1000
ARG GID=1000
RUN groupadd -g ${GID} analyst \
    && useradd -m -u ${UID} -g ${GID} analyst
USER analyst

ENV MCP_TRANSPORT=http \
    MCP_ADDR=:8090 \
    REDIS_URL=redis://redis:6379/0 \
    WIRE_KEY=analyst:wire:latest

EXPOSE 8090
CMD ["python", "-m", "app.mcp_server"]
