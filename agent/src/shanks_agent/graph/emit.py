from __future__ import annotations

from typing import Any

from langgraph.config import get_stream_writer


def emit(event: dict[str, Any]) -> None:
    """Push one UI event onto the active LangGraph custom stream.

    ``run_turn`` consumes that stream with ``astream(..., stream_mode="custom")``,
    so the event leaves the node while the node is still running.
    """
    get_stream_writer()(event)
