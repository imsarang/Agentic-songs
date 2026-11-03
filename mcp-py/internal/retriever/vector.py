import os

from qdrant_client import QdrantClient

COLLECTION = "study"


def client() -> QdrantClient:
    return QdrantClient(url=os.environ["QDRANT_URL"])


def semantic(query_vector: list[float], limit: int = 8) -> list[dict]:
    points = client().query_points(
        collection_name=COLLECTION,
        query=query_vector,
        limit=limit,
        with_payload=True,
    ).points
    return [_chunk(point) for point in points]


def _chunk(point) -> dict:
    payload = point.payload or {}
    return {
        "id": str(point.id),
        "url": payload.get("url"),
        "chunk_index": payload.get("chunk_index"),
        "body": payload.get("body") or "",
        "sha256": payload.get("sha256") or ""
    }
