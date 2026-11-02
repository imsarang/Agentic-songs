"""HTTP routes attached to the MCP server."""

from __future__ import annotations

import os

from starlette.requests import Request
from starlette.responses import JSONResponse, Response

from internal.audio_analyst import decisions
from internal.audio_analyst.ear import sample_features
from internal.audio_analyst.pipeline import decide
from internal.audio_analyst.tools.decisions import _truthy


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


async def features_route(_request: Request) -> Response:
    try:
        feat, device = sample_features()
    except Exception as e:
        return JSONResponse({"ok": False, "error": f"capture failed: {e}"}, status_code=503)
    return JSONResponse(decisions.attach({"ok": True, "device": device, "features": feat}))


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
