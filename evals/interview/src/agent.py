import json

from config import AGENT_URL
from post import open_json


def retrieve(question: str) -> list[str]:
    with open_json(AGENT_URL + "/mcp/call", {
        "name": "retrieve",
        "arguments": {"query": question, "limit": 5},
    }) as response:
        body = json.load(response)
    if "preview" not in body:
        raise RuntimeError(body.get("error") or "retrieve failed")
    return [chunk["body"] for chunk in json.loads(body["preview"])["chunks"]]


def ask(question: str) -> str:
    parts = []
    with open_json(AGENT_URL + "/turn", {"text": question, "thread_id": "eval"}) as response:
        for line in response:
            line = line.decode().strip()
            if not line.startswith("data: "):
                continue
            event = json.loads(line[6:])
            if event.get("type") == "assistant_message":
                parts.append(event.get("text") or "")
    return "\n".join(parts)
