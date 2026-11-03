import json
import os
import uuid

from redis import Redis

from internal.retriever.vector import COLLECTION, _chunk, client as qdrant

VERSION = "v1"
CUTOFF = 0.92
TTL_S = 60 * 60
PREFIX = f"retrieve:cache:{VERSION}:"

_redis = None


def _store() -> Redis:
    global _redis
    if _redis is None:
        _redis = Redis.from_url(os.getenv("REDIS_URL", "redis://127.0.0.1:6380/0"))
    return _redis


def _dot(left: list[float], right: list[float]) -> float:
    return sum(a * b for a, b in zip(left, right))


def get(vector: list[float]) -> list[dict] | None:
    try:
        store = _store()
        best_key, best_score, best = None, CUTOFF, None
        for key in store.scan_iter(f"{PREFIX}*"):
            raw = store.get(key)
            if not raw:
                continue
            row = json.loads(raw)
            score = _dot(vector, row["vector"])
            if score >= best_score:
                best_key, best_score, best = key, score, row
    except Exception:
        return None
    if best is None:
        return None
    chunks = _load(best)
    if chunks is None:
        try:
            store.delete(best_key)
        except Exception:
            pass
        return None
    for chunk in chunks:
        chunk["score"] = best_score
    return chunks


def put(vector: list[float], chunks: list[dict]) -> None:
    if not chunks:
        return
    row = {
        "vector": vector,
        "ids": [chunk["id"] for chunk in chunks],
        "sha256": {chunk["id"]: chunk.get("sha256") for chunk in chunks},
    }
    try:
        _store().set(f"{PREFIX}{uuid.uuid4()}", json.dumps(row), ex=TTL_S)
    except Exception:
        return


def _load(row: dict) -> list[dict] | None:
    points = qdrant().retrieve(COLLECTION, row["ids"], with_payload=True)
    by_id = {str(point.id): point for point in points}
    loaded = []
    for chunk_id in row["ids"]:
        point = by_id.get(chunk_id)
        payload = point.payload if point is not None else None
        if payload is None or payload.get("sha256") != row["sha256"].get(chunk_id):
            return None
        loaded.append(_chunk(point))
    return loaded
