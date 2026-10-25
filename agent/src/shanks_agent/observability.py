from __future__ import annotations

import logging
import os
import uuid

from shanks_agent.config import Settings, get_settings
from langsmith.utils import get_env_var

logger = logging.getLogger(__name__)

def new_run_id(value: str | None = None) -> uuid.UUID:
    if value:
        try:
            return uuid.UUID(value)
        except ValueError:
            logger.warning("ignoring invalid run id: %s", value)
    return uuid.uuid4()

def configure_tracing(settings: Settings | None = None) -> bool:
    settings = settings or get_settings()

    enabled = bool(settings.langsmith_tracing and settings.langsmith_api_key)

    flag = "true" if enabled else "false"

    os.environ["LANGSMITH_TRACING"] = flag
    os.environ["LANGSMITH_TRACING_V2"] = flag
    os.environ["LANGCHAIN_TRACING_V2"] = flag

    if settings.langsmith_api_key:
        os.environ["LANGSMITH_API_KEY"] = settings.langsmith_api_key
    if settings.langsmith_project:
        os.environ["LANGSMITH_PROJECT"] = settings.langsmith_project
    if settings.langsmith_endpoint:
        os.environ["LANGSMITH_ENDPOINT"] = settings.langsmith_endpoint
    get_env_var.cache_clear()

    logger.info(
        "langsmith tracing=%s project=%s",
        enabled,
        settings.langsmith_project or "default",
    )
    return enabled