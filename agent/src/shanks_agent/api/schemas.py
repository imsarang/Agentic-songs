from typing import Any

from pydantic import BaseModel, Field


class TurnRequest(BaseModel):
    text: str = Field(min_length=1)
    thread_id: str = "default"
    run_id: str | None = None
    # spotify | machine | omit → LLM classifies
    domain: str | None = None


class MCPCallRequest(BaseModel):
    name: str = Field(min_length=1)
    arguments: dict[str, Any] = Field(default_factory=dict)
