from langchain_anthropic import ChatAnthropic
from langchain_core.messages import AIMessage, BaseMessage
from langchain_openai import ChatOpenAI

from shanks_agent.config import Settings
from shanks_agent.llm.protocol import LLMClient

# Hugging Face Inference Providers (OpenAI-compatible router)
_HF_BASE_URL = "https://router.huggingface.co/v1"


class OpenAILLM(LLMClient):
    name = "openai"

    def __init__(self, settings: Settings) -> None:
        if not settings.openai_api_key:
            raise ValueError("OPENAI_API_KEY is empty")
        self._model = ChatOpenAI(
            api_key=settings.openai_api_key,
            model=settings.openai_model,
            temperature=0,
        )

    async def acomplete(self, messages: list[BaseMessage]) -> AIMessage:
        out = await self._model.ainvoke(messages)
        return out if isinstance(out, AIMessage) else AIMessage(content=str(out.content))


class AnthropicLLM(LLMClient):
    name = "anthropic"

    def __init__(self, settings: Settings) -> None:
        if not settings.anthropic_api_key:
            raise ValueError("ANTHROPIC_API_KEY is empty")
        self._model = ChatAnthropic(
            api_key=settings.anthropic_api_key,
            model=settings.anthropic_model,
            temperature=0,
        )

    async def acomplete(self, messages: list[BaseMessage]) -> AIMessage:
        out = await self._model.ainvoke(messages)
        return out if isinstance(out, AIMessage) else AIMessage(content=str(out.content))


class HuggingFaceLLM(LLMClient):
    """Chat via HF Inference Providers router (not the old serverless text-gen API)."""

    name = "huggingface"

    def __init__(self, settings: Settings) -> None:
        if not settings.huggingface_api_key:
            raise ValueError("HUGGINGFACE_API_KEY is empty")
        model = settings.huggingface_model
        if settings.huggingface_provider:
            model = f"{model}:{settings.huggingface_provider}"
        self._model = ChatOpenAI(
            api_key=settings.huggingface_api_key,
            base_url=_HF_BASE_URL,
            model=model,
            temperature=0,
        )

    async def acomplete(self, messages: list[BaseMessage]) -> AIMessage:
        out = await self._model.ainvoke(messages)
        return out if isinstance(out, AIMessage) else AIMessage(content=str(out.content))


class LocalLLMStudio(LLMClient):
    """Any OpenAI-compatible local server (LM Studio, Ollama, llama.cpp, vLLM)."""

    name = "local"

    def __init__(self, settings: Settings) -> None:
        if not settings.local_base_url:
            raise ValueError("LOCAL_BASE_URL is empty")
        if not settings.local_model:
            raise ValueError("LOCAL_MODEL is empty")
        self._model = ChatOpenAI(
            api_key=settings.local_api_key or "local",
            base_url=settings.local_base_url.rstrip("/"),
            model=settings.local_model,
            temperature=0.7,
        )

    async def acomplete(self, messages: list[BaseMessage]) -> AIMessage:
        out = await self._model.ainvoke(messages)
        return out if isinstance(out, AIMessage) else AIMessage(content=str(out.content))


PROVIDERS: dict[str, type[LLMClient]] = {
    "openai": OpenAILLM,
    "anthropic": AnthropicLLM,
    "huggingface": HuggingFaceLLM,
    "local": LocalLLMStudio,
}
