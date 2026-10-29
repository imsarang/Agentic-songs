import json
import os
from pathlib import Path

import numpy as np

os.environ.setdefault("S3_ENDPOINT", "http://127.0.0.1:9000")
os.environ.setdefault("S3_BUCKET", "shanks-raw")
os.environ.setdefault("S3_ACCESS_KEY", "shanks")
os.environ.setdefault("S3_SECRET_KEY", "shanks-secret")
os.environ.setdefault("S3_REGION", "us-east-1")
os.environ.setdefault("EMBED_MODEL_PATH", str(Path.home() / ".local/models/bge-large-en-v1.5"))

import sys

sys.path.insert(0, str(Path(__file__).resolve().parent / "src"))

from worker.chunk import Chunker
from worker.embed import Embedder
from worker.objects import get_object

SIZES = (128, 256, 384, 510)
OVERLAP = 50
K = 5
QUERY = "Represent this sentence for searching relevant passages: "
TEXT_KEY = (
    "text/hello-interview-rate-limitter/"
    "f94571551026676943f158f0cfdfc5c4987bdfdc85fd07f08b97fee1bfcf9678.txt"
)


def relevant(chunks: list[str], needles: list[str]) -> set[int]:
    folded = [chunk.lower() for chunk in chunks]
    need = [needle.lower() for needle in needles]
    return {i for i, chunk in enumerate(folded) if all(needle in chunk for needle in need)}


def top_k(matrix: np.ndarray, query: np.ndarray, k: int) -> list[int]:
    scores = matrix @ query
    k = min(k, len(scores))
    picked = np.argpartition(scores, -k)[-k:]
    order = picked[np.argsort(scores[picked])[::-1]]
    return [int(i) for i in order]


def evaluate(chunks: list[str], matrix: np.ndarray, questions: list[dict], queries: np.ndarray) -> dict:
    accuracy = precision = recall = 0.0
    for row, item in zip(queries, questions):
        gold = relevant(chunks, item["contains"])
        got = set(top_k(matrix, row, K))
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


def main() -> None:
    questions = json.loads((Path(__file__).parent / "questions.json").read_text())
    text = get_object(TEXT_KEY).decode()
    embedder = Embedder()
    print(f"device {embedder.device}")
    queries = np.asarray(
        embedder.embed([QUERY + item["question"] for item in questions]),
        dtype=np.float32,
    )
    rows = []
    for size in SIZES:
        chunker = Chunker(size=size, overlap=OVERLAP)
        chunker.tokenizer.model_max_length = 1_000_000
        chunks = chunker.split(text)
        matrix = np.asarray(embedder.embed(chunks), dtype=np.float32)
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
