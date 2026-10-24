from abc import ABC, abstractmethod

from langchain_core.messages import AIMessage, BaseMessage


class LLMClient(ABC):
    """Contract for every LLM backend. Subclass this to add a provider."""

    name: str

    @abstractmethod
    async def acomplete(self, messages: list[BaseMessage]) -> AIMessage:
        """Return one assistant message for the given chat history."""
        ...
