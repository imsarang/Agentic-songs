"""audio_decisions tool."""

from __future__ import annotations

import json

from internal.audio_analyst.pipeline import decide


def _truthy(v: str | bool | None) -> bool:
    if isinstance(v, bool):
        return v
    if v is None:
        return False
    return str(v).strip().lower() in ("1", "true", "yes", "y", "finalize", "end")


def audio_decisions(trackId: str, finalize: str = "false") -> str:
    """trackId: Spotify track id or URI. finalize: true when paused/ended."""
    result = decide(trackId, finalize=_truthy(finalize))
    result.pop("device", None)
    return json.dumps(result, indent=2)
