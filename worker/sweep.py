import json
import os
import sys
from pathlib import Path

import numpy as np

os.environ.setdefault("S3_ENDPOINT", "http://127.0.0.1:9000")
os.environ.setdefault("S3_BUCKET", "shanks-raw")
os.environ.setdefault("S3_ACCESS_KEY", "shanks")
os.environ.setdefault("S3_SECRET_KEY", "shanks-secret")
os.environ.setdefault("S3_REGION", "us-east-1")
os.environ.setdefault("EMBED_URL", "http://127.0.0.1:8091/embed")
os.environ.setdefault("EMBED_MODEL_PATH", str(Path.home() / ".local/models/bge-large-en-v1.5"))

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "worker" / "src"))
sys.path.insert(0, str(ROOT / "mcp-py"))

from internal.retriever.fuse import _rrf
from internal.retriever.keyword import _bm25
from worker.chunk import Chunker
from worker.embed import embed
from worker.objects import get_object

SIZES = (128, 256, 384, 510)
OVERLAP = 50
K = 5
POOL = 8
SOURCE_ID = "hello-interview-rate-limitter"


def relevant(chunks: list[str], needles: list[str]) -> set[str]:
    need = [needle.lower() for needle in needles]
    return {
        str(i)
        for i, chunk in enumerate(chunks)
        if all(needle in chunk.lower() for needle in need)
    }


def semantic(chunks: list[str], matrix: np.ndarray, query: np.ndarray) -> list[dict]:
    scores = matrix @ query
    k = min(POOL, len(scores))
    picked = np.argpartition(scores, -k)[-k:]
    order = picked[np.argsort(scores[picked])[::-1]]
    return [{"id": str(int(i)), "body": chunks[int(i)]} for i in order]


def keyword(query: str, chunks: list[str]) -> list[dict]:
    ranked = sorted(enumerate(_bm25(query, chunks)), key=lambda row: row[1], reverse=True)
    return [
        {"id": str(i), "body": chunks[i]}
        for i, score in ranked[:POOL]
        if score > 0
    ]


def evaluate(chunks: list[str], matrix: np.ndarray, questions: list[dict], queries: np.ndarray) -> dict:
    accuracy = precision = recall = 0.0
    for row, item in zip(queries, questions):
        gold = relevant(chunks, item["contains"])
        got = {chunk["id"] for chunk in _rrf([
            semantic(chunks, matrix, row),
            keyword(item["question"], chunks),
        ])[:K]}
        hit = gold & got
        accuracy += 1.0 if hit else 0.0
        precision += len(hit) / K
        recall += (len(hit) / len(gold)) if gold else 0.0
    n = len(questions)
    return {
        "accuracy": accuracy / n,
        "precision": precision / n,
        "recall": recall / n,
    }


def text_key(source_id: str) -> str:
    import boto3

    client = boto3.client(
        "s3",
        endpoint_url=os.environ["S3_ENDPOINT"],
        aws_access_key_id=os.environ["S3_ACCESS_KEY"],
        aws_secret_access_key=os.environ["S3_SECRET_KEY"],
        region_name=os.environ.get("S3_REGION", "us-east-1"),
    )
    found = client.list_objects_v2(
        Bucket=os.environ.get("S3_BUCKET", "shanks-raw"),
        Prefix=f"text/{source_id}/",
    )
    keys = [item["Key"] for item in found.get("Contents") or [] if item["Key"].endswith(".txt")]
    if not keys:
        raise SystemExit(f"no text object for {source_id}")
    return keys[0]


def main() -> None:
    questions = json.loads((Path(__file__).parent / "questions.json").read_text())
    text = get_object(text_key(SOURCE_ID)).decode()
    queries = np.asarray(
        embed([item["question"] for item in questions], "query"),
        dtype=np.float32,
    )
    rows = []
    for size in SIZES:
        chunker = Chunker(size=size, overlap=OVERLAP)
        chunker.tokenizer.model_max_length = 1_000_000
        chunks = chunker.split(text)
        matrix = np.asarray(embed(chunks, "document"), dtype=np.float32)
        scores = evaluate(chunks, matrix, questions, queries)
        rows.append((size, len(chunks), scores))
        print(
            f"size {size} chunks {len(chunks)} "
            f"accuracy {scores['accuracy']:.2f} "
            f"precision {scores['precision']:.2f} "
            f"recall {scores['recall']:.2f}"
        )
    size, count, scores = max(
        rows,
        key=lambda row: (row[2]["accuracy"], row[2]["recall"], row[2]["precision"]),
    )
    print(
        f"best size {size} tokens, {count} chunks, "
        f"accuracy {scores['accuracy']:.2f} "
        f"precision {scores['precision']:.2f} "
        f"recall {scores['recall']:.2f}"
    )


if __name__ == "__main__":
    main()
