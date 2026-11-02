import json
import os
import sys
from pathlib import Path

os.environ.setdefault("QDRANT_URL", "http://127.0.0.1:6333")
os.environ.setdefault("EMBED_URL", "http://127.0.0.1:8091/embed")

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "mcp-py"))

from internal.retriever.fuse import retrieve
from internal.retriever.vector import COLLECTION, _chunk, client

K = 5


def load_chunks() -> list[dict]:
    found = []
    offset = None
    qdrant = client()
    while True:
        points, offset = qdrant.scroll(
            collection_name=COLLECTION,
            limit=128,
            offset=offset,
            with_payload=True,
        )
        found.extend(_chunk(point) for point in points)
        if offset is None:
            break
    return found


def gold_ids(chunks: list[dict], needles: list[str]) -> set[str]:
    need = [needle.lower() for needle in needles]
    return {
        chunk["id"]
        for chunk in chunks
        if all(needle in chunk["body"].lower() for needle in need)
    }


def main() -> None:
    questions = json.loads((Path(__file__).parent / "questions.json").read_text())
    chunks = load_chunks()
    accuracy = precision = recall = 0.0
    for item in questions:
        gold = gold_ids(chunks, item["contains"])
        got = {chunk["id"] for chunk in retrieve(item["question"], K)}
        hit = gold & got
        accuracy += 1.0 if hit else 0.0
        precision += len(hit) / K
        recall += (len(hit) / len(gold)) if gold else 0.0
        print(f"{'hit' if hit else 'miss'} gold={len(gold)} {item['question'][:70]}")
    n = len(questions)
    print(
        f"accuracy {accuracy / n:.2f} "
        f"precision {precision / n:.2f} "
        f"recall {recall / n:.2f}"
    )


if __name__ == "__main__":
    main()