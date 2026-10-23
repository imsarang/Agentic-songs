from __future__ import annotations

import json
from contextlib import asynccontextmanager
from typing import Any, AsyncIterator

from mcp import ClientSession

from shanks_agent.config import get_settings

try:
    from mcp.client.streamable_http import streamable_http_client as _http_client
except ImportError:
    from mcp.client.streamable_http import streamablehttp_client as _http_client  # type: ignore


def format_mcp_error(exc: BaseException) -> str:
    """Unwrap ExceptionGroup / TaskGroup noise from the MCP SDK."""
    if isinstance(exc, BaseExceptionGroup):
        return "; ".join(format_mcp_error(e) for e in exc.exceptions)
    return f"{type(exc).__name__}: {exc}"



class MCPSession:
    def __init__(self, url: str | None = None) -> None:
        self._url = url or get_settings().mcp_url

    async def list_tools(self) -> list[str]:
        async with _connected(self._url) as session:
            result = await session.list_tools()
            return [t.name for t in result.tools]

    async def call_tool(self, name: str, arguments: dict[str, Any] | None = None) -> str:
        args = arguments or {}
        async with _connected(self._url) as session:
            result = await session.call_tool(name, args)
            return _result_text(result)


@asynccontextmanager
async def _connected(url: str) -> AsyncIterator[ClientSession]:
    async with _http_client(url) as streams:
        read, write = streams[0], streams[1]
        async with ClientSession(read, write) as session:
            await session.initialize()
            yield session


def _result_text(result: Any) -> str:
    parts: list[str] = []
    for block in getattr(result, "content", []) or []:
        text = getattr(block, "text", None)
        if text:
            parts.append(text)
        else:
            parts.append(str(block))
    if parts:
        return "\n".join(parts)
    if getattr(result, "isError", False):
        return json.dumps({"error": True, "raw": str(result)})
    return str(result)
