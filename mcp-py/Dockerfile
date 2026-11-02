FROM python:3.12-slim

WORKDIR /app
COPY mcp-py/requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --timeout 300 --retries 10 -r requirements.txt
COPY mcp-py ./mcp-py

ARG UID=1000
ARG GID=1000
RUN groupadd -g ${GID} analyst \
    && useradd -m -u ${UID} -g ${GID} analyst
USER analyst

ENV MCP_TRANSPORT=http \
    MCP_ADDR=:8093 \
    PYTHONPATH=/app/mcp-py

EXPOSE 8093
CMD ["python", "/app/mcp-py/cmd/retriever/main.py"]
