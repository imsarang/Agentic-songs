import logging
import time

from internal.retriever.cache import get, put
from internal.retriever.embed import embed
from internal.retriever.keyword import keyword
from internal.retriever.series import write
from internal.retriever.vector import semantic

logger = logging.getLogger(__name__)


def retrieve(query: str, limit: int = 5, run_id: str = "") -> list[dict]:
    started = time.perf_counter()
    vector = embed([query], "query")[0]
    picked = get(vector)
    hit = picked is not None
    if not hit:
        picked = _rrf([semantic(vector, limit=8), keyword(query, limit=8)])[:limit]
        if picked:
            put(vector, picked)
    elapsed_ms = (time.perf_counter() - started) * 1000
    logger.info(
        "retrieve run_id=%s query=%r chunk_ids=%s scores=%s empty=%s cache=%s latency_ms=%.0f",
        run_id,
        query,
        [chunk["id"] for chunk in picked],
        [round(chunk["score"], 4) for chunk in picked],
        not picked,
        hit,
        elapsed_ms,
    )
    write(elapsed_ms, not picked, len(picked))
    return picked


def _rrf(rankings: list[list[dict]], k: int = 60) -> list[dict]:
    scores: dict[str, float] = {}
    chunks: dict[str, dict] = {}
    for ranking in rankings:
        for rank, chunk in enumerate(ranking):
            chunk_id = chunk["id"]
            chunks[chunk_id] = chunk
            scores[chunk_id] = scores.get(chunk_id, 0.0) + 1 / (k + rank + 1)
    ordered = sorted(scores, key=scores.get, reverse=True)
    return [{**chunks[chunk_id], "score": scores[chunk_id]} for chunk_id in ordered]