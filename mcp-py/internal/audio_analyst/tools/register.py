"""Register audio-analyst tools. Same role as mcp/internal/*/tools/register.go."""

from __future__ import annotations

from mcp.server.fastmcp import FastMCP

from internal.audio_analyst.tools.decisions import audio_decisions
from internal.audio_analyst.tools.routes import decisions_route, features_route, healthz

_AUDIO_DECISIONS = (
    "One live-audio sample for the current Spotify track. Returns in ~WINDOW_MS "
    "(a few seconds) — it does NOT wait for the song to finish. "
    "CRITICAL UX: as soon as this tool returns, you MUST immediately write a short "
    "user-visible message with summary_bullets (sample i/n). Do NOT wait for more "
    "samples, do NOT batch until the song ends, do NOT stay silent between calls. "
    "Then wait ~15s (or wait_ms), call again, message again — up to 5 samples. "
    "On pause/end: finalize=true once, message the weighted avg. "
    "If silence/ok=false: say capture failed; do not invent song lore. "
    "Play must not fail if this tool errors."
)


def register(mcp: FastMCP) -> None:
    mcp.tool(name="audio_decisions", description=_AUDIO_DECISIONS)(audio_decisions)
    mcp.custom_route("/healthz", methods=["GET"])(healthz)
    mcp.custom_route("/features", methods=["GET"])(features_route)
    mcp.custom_route("/decisions", methods=["GET"])(decisions_route)
