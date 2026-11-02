import json

from mcp.server.fastmcp import FastMCP

from internal.retriever.fuse import retrieve


def register(mcp: FastMCP) -> None:
    mcp.tool(
        name="retrieve",
        description="Retrieve study chunks for a question. Returns the full text of each chunk.",
    )(retrieve_chunks)


def retrieve_chunks(query: str, limit: int = 5, run_id: str = "") -> str:
    chunks = retrieve(query, limit, run_id)
    return json.dumps({"query": query, "chunks": chunks})