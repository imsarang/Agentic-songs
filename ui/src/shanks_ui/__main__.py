import asyncio

from shanks_ui.app import run_repl


def main() -> None:
    asyncio.run(run_repl())


if __name__ == "__main__":
    main()