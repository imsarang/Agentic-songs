"""Retriever MCP entrypoint."""

import logging
import sys
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s %(message)s")

_MCP_PY = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(_MCP_PY))

from internal.retriever.tools.register import register
from internal.transport.serve import serve


def main() -> None:
    import os
    os.environ.setdefault("MCP_ADDR", ":8093")
    serve("retriever", register)


if __name__ == "__main__":
    main()
