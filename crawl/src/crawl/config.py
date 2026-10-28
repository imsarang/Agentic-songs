from pathlib import Path
import yaml
from pydantic_settings import BaseSettings, SettingsConfigDict
from crawl.models import Source

class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        extra="ignore"
    )

    firecrawl_api_url: str = "https://api.firecrawl.dev"
    firecrawl_api_key: str
    sources_path: Path = Path("config/sources.yaml")
    s3_endpoint: str
    s3_bucket: str = "shanks-raw"
    s3_access_key: str
    s3_secret_key: str
    s3_region: str = "us-east-1"
    celery_broker_url: str = "redis://127.0.0.1:6380/1"


def load_sources(path: Path) -> list[Source]:
    raw = yaml.safe_load(path.read_text())
    return [Source.model_validate(row) for row in raw["sources"]]