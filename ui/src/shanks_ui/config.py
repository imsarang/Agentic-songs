from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    agent_turn_url: str = "http://127.0.0.1:8001/turn"
    thread_id: str = "default"


@lru_cache
def get_settings() -> Settings:
    return Settings()