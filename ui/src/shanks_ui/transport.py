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
    timeout: float = 180.0,
) -> AsyncIterator[dict[str, Any]]:
    """POST /turn and yield parsed SSE `data:` JSON objects."""
    async with httpx.AsyncClient(timeout=timeout) as client:
        async with client.stream(
            "POST",
            url,
            json={"text": text, "thread_id": thread_id},
            headers={"Accept": "text/event-stream"},
        ) as resp:
            resp.raise_for_status()
            buf = ""
            async for chunk in resp.aiter_text():
                buf += chunk
                while "\n\n" in buf:
                    block, buf = buf.split("\n\n", 1)
                    for line in block.splitlines():
                        if not line.startswith("data:"):
                            continue
                        raw = line[5:].strip()
                        if not raw:
                            continue
                        yield json.loads(raw)
