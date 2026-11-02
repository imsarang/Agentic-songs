"""Audio-analyst MCP entrypoint.

MCP_TRANSPORT=stdio (default) | http
MCP_ADDR=:8090
"""

import sys
from pathlib import Path

_MCP_PY = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(_MCP_PY))

from internal.audio_analyst.tools.register import register
from internal.transport.serve import serve


def main() -> None:
    serve("audio-analyst", register)


if __name__ == "__main__":
    main()
