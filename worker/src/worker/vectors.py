import os
import uuid

from qdrant_client import QdrantClient
from qdrant_client.models import (
    Distance,
    FieldCondition,
    Filter,
    MatchValue,
    PointStruct,
    TextIndexParams,
    TokenizerType,
    VectorParams,
)

_COLLECTION = "study"
_NAMESPACE = uuid.UUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8")


class VectorStore:
    def __init__(self) -> None:
        self.client = QdrantClient(url=os.environ["QDRANT_URL"])
        self._ensure()

    def _ensure(self) -> None:
        if not self.client.collection_exists(_COLLECTION):
            
            self.client.create_collection(
                collection_name=_COLLECTION,
                vectors_config=VectorParams(size=1024, distance=Distance.COSINE),
            )
        
            self.client.create_payload_index(
                collection_name=_COLLECTION,
                field_name="body",
                field_schema=TextIndexParams(
                    type="text",
                    tokenizer=TokenizerType.WORD,
                    min_token_len=2,
                    max_token_len=20,
                    lowercase=True,
                ),
            )
        schema = self.client.get_collection(_COLLECTION).payload_schema or {}
        if "sha256" not in schema:
            self.client.create_payload_index(
                collection_name = _COLLECTION,
                field_name="sha256",
                field_schema="keyword"
            )

    def has_sha256(self, digest: str) -> bool:
        points, _ = self.client.scroll(
            collection_name=_COLLECTION,
            scroll_filter=Filter(
                must=[FieldCondition(key="sha256", match=MatchValue(value=digest))]
            ),
            limit=1,
            with_payload=False,
        )
        return bool(points)

    def upsert(self, message: dict, parts: list[str], vectors: list[list[float]]) -> int:
        points = [
            PointStruct(
                id=str(uuid.uuid5(_NAMESPACE, f"{message['sha256']}:{index}")),
                vector=vector,
                payload={
                    "source_id": message["source_id"],
                    "sha256": message["sha256"],
                    "url": message["url"],
                    "chunk_index": index,
                    "body": part,
                },
            )
            for index, (part, vector) in enumerate(zip(parts, vectors))
        ]
        self.client.upsert(collection_name=_COLLECTION, points=points)
        return len(points)