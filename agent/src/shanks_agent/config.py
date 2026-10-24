from functools import lru_cache
from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict

_AGENT_ROOT = Path(__file__).resolve().parents[2]
_ENV_FILE = _AGENT_ROOT / ".env"



class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=str(_ENV_FILE) if _ENV_FILE.exists() else ".env",
        extra="ignore",
    )

    agent_host: str = "0.0.0.0"
    agent_port: int = 8000
    mcp_url: str = "http://127.0.0.1:8080/mcp"

    llm_providers: str = "openai,anthropic,huggingface"
    openai_api_key: str = ""
    openai_model: str = "gpt-4o-mini"
    anthropic_api_key: str = ""
    anthropic_model: str = "claude-sonnet-4-20250514"
    huggingface_api_key: str = ""
    # Must be a model that has Inference Providers on the Hub (check model card).
    huggingface_model: str = "mistralai/Mistral-7B-Instruct-v0.2"
    # Optional pin, e.g. together | fireworks-ai | sambanova | hf-inference | groq
    huggingface_provider: str = ""

    # LM Studio default: http://127.0.0.1:1234/v1 — Ollama often :11434/v1
    local_base_url: str = "http://localhost:1234/v1"
    local_api_key: str = "local"
    local_model: str = "qwen2.5-7b-instruct-1m"

    def provider_names(self) -> list[str]:
        return [p.strip().lower() for p in self.llm_providers.split(",") if p.strip()]

@lru_cache
def get_settings() -> Settings:
    return Settings()

