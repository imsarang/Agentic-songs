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

    agent_host: str
    agent_port: int
    mcp_url: str
    machine_mcp_url: str
    google_mcp_url: str = "http://127.0.0.1:8004/mcp"

    llm_providers: str
    openai_api_key: str
    openai_model: str
    anthropic_api_key: str
    anthropic_model: str
    huggingface_api_key: str
    huggingface_model: str
    huggingface_provider: str

    local_base_url: str
    local_api_key: str
    local_model: str

    langsmith_tracing: bool
    langsmith_api_key: str
    langsmith_project: str
    langsmith_endpoint: str

    def provider_names(self) -> list[str]:
        return [p.strip().lower() for p in self.llm_providers.split(",") if p.strip()]

@lru_cache
def get_settings() -> Settings:
    return Settings()

