from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any

import httpx


async def stream_turn(
    url: str,
    text: str,
    *,
    thread_id: str = "default",
    run_id: str | None = None,
    timeout: float = 180.0,
) -> AsyncIterator[dict[str, Any]]:
    """POST /turn and yield parsed SSE `data:` JSON objects."""
    payload: dict[str, Any] = {"text": text, "thread_id": thread_id}
    if run_id:
        payload["run_id"] = run_id
    async with httpx.AsyncClient(timeout=timeout) as client:
        async with client.stream(
            "POST",
            url,
            json=payload,
            headers={"Accept": "text/event-stream"},
        ) as resp:
            resp.raise_for_status()
            async for line in resp.aiter_lines():
                if not line.startswith("data:"):
                    continue
                raw = line[5:].strip()
                if not raw:
                    continue
                yield json.loads(raw)
