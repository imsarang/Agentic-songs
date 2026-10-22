"""Audio-analyst MCP.

MCP_TRANSPORT=stdio (default) | http
MCP_ADDR=:8090                 (http listen, same idea as Spotify MCP)

  stdio → Cursor/host spawns this process (capture on host PipeWire)
  http  → container/long-running server at http://127.0.0.1:8090/mcp
"""

from __future__ import annotations

import json
import os

os.environ.setdefault("JACK_NO_START_SERVER", "1")

from mcp.server.fastmcp import FastMCP
from mcp.server.transport_security import TransportSecuritySettings
from starlette.requests import Request
from starlette.responses import JSONResponse, Response

from app import decisions
from app.ear import sample_features
from app.pipeline import decide

# DNS rebinding off so Docker / non-localhost Host headers work (like Spotify
# server.WithDisableLocalhostProtection).
_mcp_kwargs: dict = {
    "host": "0.0.0.0",
    "port": 8090,
    "streamable_http_path": "/mcp",
    "stateless_http": True,
    "transport_security": TransportSecuritySettings(
        enable_dns_rebinding_protection=False,
    ),
}

mcp = FastMCP("audio-analyst", **_mcp_kwargs)


def _truthy(v: str | bool | None) -> bool:
    if isinstance(v, bool):
        return v
    if v is None:
        return False
    return str(v).strip().lower() in ("1", "true", "yes", "y", "finalize", "end")


def _parse_addr(addr: str) -> tuple[str, int]:
    """':8090' or '0.0.0.0:8090' → host, port."""
    a = (addr or ":8090").strip()
    if a.startswith(":"):
        return "0.0.0.0", int(a[1:])
    if ":" in a:
        host, port_s = a.rsplit(":", 1)
        return (host or "0.0.0.0"), int(port_s)
    return "0.0.0.0", int(a)


@mcp.tool(
    name="audio_decisions",
    description=(
        "One live-audio sample for the current Spotify track. Returns in ~WINDOW_MS "
        "(a few seconds) — it does NOT wait for the song to finish. "
        "CRITICAL UX: as soon as this tool returns, you MUST immediately write a short "
        "user-visible message with summary_bullets (sample i/n). Do NOT wait for more "
        "samples, do NOT batch until the song ends, do NOT stay silent between calls. "
        "Then wait ~15s (or wait_ms), call again, message again — up to 5 samples. "
        "On pause/end: finalize=true once, message the weighted avg. "
        "If silence/ok=false: say capture failed; do not invent song lore. "
        "Play must not fail if this tool errors."
    ),
)
def audio_decisions(trackId: str, finalize: str = "false") -> str:
    """trackId: Spotify track id or URI. finalize: true when paused/ended."""
    result = decide(trackId, finalize=_truthy(finalize))
    result.pop("device", None)
    return json.dumps(result, indent=2)


@mcp.custom_route("/healthz", methods=["GET"])
async def healthz(_request: Request) -> Response:
    return JSONResponse(
        {
            "ok": True,
            "service": "audio-analyst",
            "transport": os.getenv("MCP_TRANSPORT", "stdio"),
            "wire_key": os.getenv("WIRE_KEY", "analyst:wire:latest"),
            "addr": os.getenv("MCP_ADDR", ":8090"),
        }
    )


@mcp.custom_route("/features", methods=["GET"])
async def features_route(_request: Request) -> Response:
    try:
        feat, device = sample_features()
    except Exception as e:
        return JSONResponse({"ok": False, "error": f"capture failed: {e}"}, status_code=503)
    return JSONResponse(decisions.attach({"ok": True, "device": device, "features": feat}))


@mcp.custom_route("/decisions", methods=["GET"])
async def decisions_route(request: Request) -> Response:
    track_id = request.query_params.get("trackId") or ""
    finalize = _truthy(request.query_params.get("finalize"))
    result = decide(track_id, finalize=finalize)
    if not result.get("ok"):
        err = result.get("error", "unknown error")
        code = 400 if "required" in err or "no samples" in err else 503
        return JSONResponse({"ok": False, "error": err}, status_code=code)
    result.pop("device", None)
    return JSONResponse(result)


def main() -> None:
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


if __name__ == "__main__":
    main()
