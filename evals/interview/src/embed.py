import json

from langchain_core.embeddings import Embeddings

from config import EMBED_URL
from post import open_json


class EmbedServer(Embeddings):
    def _vectors(self, texts: list[str], kind: str) -> list[list[float]]:
        with open_json(EMBED_URL, {"texts": texts, "kind": kind}) as response:
            return json.load(response)["vectors"]

    def embed_documents(self, texts: list[str]) -> list[list[float]]:
        return self._vectors(texts, "document")

    def embed_query(self, text: str) -> list[float]:
        return self._vectors([text], "query")[0]
