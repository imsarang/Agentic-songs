import json
import os
import urllib.request


def embed(texts: list[str], kind: str) -> list[list[float]]:
    raw = json.dumps({"texts": texts, "kind": kind}).encode()
    req = urllib.request.Request(
        os.environ.get("EMBED_URL", "http://127.0.0.1:8091/embed"),
        data=raw,
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req) as res:
        return json.load(res)["vectors"]
