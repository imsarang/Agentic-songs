from __future__ import annotations

from collections.abc import Sequence

from langchain_core.messages import AIMessage, BaseMessage

from shanks_agent.config import Settings, get_settings
from shanks_agent.llm.protocol import LLMClient
from shanks_agent.llm.providers import PROVIDERS


class FailoverLLM:
    def __init__(self, clients: Sequence[LLMClient]) -> None:
        if not clients:
            raise ValueError("at least one LLM client required")
        self._clients = list(clients)

    async def acomplete(self, messages: list[BaseMessage]) -> AIMessage:
        errors: list[str] = []
        for client in self._clients:
            try:
                return await client.acomplete(messages)
            except Exception as exc:  # noqa: BLE001
                errors.append(f"{client.name}: {exc}")
        raise RuntimeError("all LLM providers failed: " + "; ".join(errors))


def build_failover_llm(settings: Settings | None = None) -> FailoverLLM:
    settings = settings or get_settings()
    clients: list[LLMClient] = []
    for name in settings.provider_names():
        cls = PROVIDERS.get(name)
        if cls is None:
            raise ValueError(f"unknown LLM provider: {name}")
        try:
            clients.append(cls(settings))
        except Exception:
            continue  # e.g. empty API key — try the next provider
    return FailoverLLM(clients)
