"""MCP transport. Does not know the tools."""

from __future__ import annotations

import os
from collections.abc import Callable

from mcp.server.fastmcp import FastMCP
from mcp.server.transport_security import TransportSecuritySettings

_mcp_kwargs: dict = {
    "host": "0.0.0.0",
    "port": 8090,
    "streamable_http_path": "/mcp",
    "stateless_http": True,
    "transport_security": TransportSecuritySettings(
        enable_dns_rebinding_protection=False,
    ),
}


def _parse_addr(addr: str) -> tuple[str, int]:
    """':8090' or '0.0.0.0:8090' → host, port."""
    a = (addr or ":8090").strip()
    if a.startswith(":"):
        return "0.0.0.0", int(a[1:])
    if ":" in a:
        host, port_s = a.rsplit(":", 1)
        return (host or "0.0.0.0"), int(port_s)
    return "0.0.0.0", int(a)


def serve(name: str, register: Callable[[FastMCP], None]) -> None:
    mcp = FastMCP(name, **_mcp_kwargs)
    register(mcp)

    transport = os.getenv("MCP_TRANSPORT", "stdio").strip().lower() or "stdio"
    if transport == "http":
        host, port = _parse_addr(os.getenv("MCP_ADDR", ":8090"))
        mcp.settings.host = host
        mcp.settings.port = port
        mcp.settings.stateless_http = True
        print(f"mcp http transport addr={host}:{port} path=/mcp", flush=True)
        mcp.run(transport="streamable-http")
    else:
        print("mcp stdio transport", flush=True)
        mcp.run(transport="stdio")
