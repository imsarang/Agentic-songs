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


@lru_cache
def get_settings() -> Settings:
    return Settings()
