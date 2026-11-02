import math
import re

from internal.retriever.vector import COLLECTION, _chunk, client

from qdrant_client.models import FieldCondition, Filter, MatchText

def keyword(query: str, limit: int = 8) -> list[dict]:
    terms = _tokens(query)
    if not terms:
        return []
    chunks = _scroll(terms)
    if not chunks:
        return []
    
    scores = _bm25(query, [chunk["body"] for chunk in chunks])
    ranked = sorted(zip(scores, chunks), key = lambda row: row[0], reverse=True)
    return [
        chunk for score, chunk in ranked[:limit] if score > 0
    ]

def _scroll(terms: list[str]) -> list[dict]:

    query_filter = Filter(
        should=[
            FieldCondition(key="body", match=MatchText(text=term))
            for term in terms
        ]
    )

    found = []
    offset = None
    qdrant = client()
    while True:
        points, offset = qdrant.scroll(
            collection_name=COLLECTION,
            limit=128,
            offset=offset,
            with_payload=True,
            scroll_filter=query_filter
        )
        found.extend(_chunk(point) for point in points)
        if offset is None:
            break
    return found


def _tokens(text: str) -> list[str]:
    return re.findall(r"[a-z0-9]{2,}", text.lower())


def _bm25(query: str, docs: list[str], k1: float = 1.5, b: float = 0.75) -> list[float]:
    query_terms = _tokens(query)
    tokenized = [_tokens(doc) for doc in docs]
    count = len(tokenized)
    avg = sum(len(tokens) for tokens in tokenized) / count
    df: dict[str, int] = {}
    for tokens in tokenized:
        for term in set(tokens):
            df[term] = df.get(term, 0) + 1
    scores = []
    for tokens in tokenized:
        tf: dict[str, int] = {}
        for term in tokens:
            tf[term] = tf.get(term, 0) + 1
        length = len(tokens) or 1
        score = 0.0
        for term in query_terms:
            if term not in tf:
                continue
            idf = math.log(1 + (count - df[term] + 0.5) / (df[term] + 0.5))
            freq = tf[term]
            score += idf * (freq * (k1 + 1)) / (freq + k1 * (1 - b + b * length / avg))
        scores.append(score)
    return scores