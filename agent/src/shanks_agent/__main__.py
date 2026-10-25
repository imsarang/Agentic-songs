import uvicorn

from shanks_agent.config import get_settings
from shanks_agent.logging_setup import setup_logging
from shanks_agent.observability import configure_tracing


def main() -> None:
    setup_logging()
    settings = get_settings()
    configure_tracing(settings)
    uvicorn.run(
        "shanks_agent.api.app:app",
        host=settings.agent_host,
        port=settings.agent_port,
        reload=False,
    )


if __name__ == "__main__":
    main()