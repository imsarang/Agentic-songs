import uvicorn

from shanks_agent.config import get_settings


def main() -> None:
    settings = get_settings()
    uvicorn.run(
        "shanks_agent.api.app:app",
        host=settings.agent_host,
        port=settings.agent_port,
        reload=False,
    )


if __name__ == "__main__":
    main()