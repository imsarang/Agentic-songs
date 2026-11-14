import json
import os
import urllib.request

from infra.ports import Embeddings


class Embedder(Embeddings):
    def __init__(self) -> None:
        self._url = os.environ.get("EMBED_URL", "http://127.0.0.1:8091/embed")

    def embed(self, texts: list[str], kind: str) -> list[list[float]]:
        
        # constructing the actual request
        raw = json.dumps({"texts": texts, "kind": kind}).encode()
        request = urllib.request.Request(
            self._url,
            data=raw,
            headers={"Content-Type": "application/json"},
        )

        # making the actual request
        with urllib.request.urlopen(request) as response:
            vectors = json.load(response)["vectors"]
        for vector in vectors:
            if len(vector) != 1024:
                raise ValueError(f"embedding length {len(vector)}")
        return vectors