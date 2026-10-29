import os

import torch
from sentence_transformers import SentenceTransformer


class Embedder:
    def __init__(self) -> None:
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        kwargs = {"torch_dtype": torch.float16} if self.device == "cuda" else {}
        self.model = SentenceTransformer(
            os.environ["EMBED_MODEL_PATH"],
            device=self.device,
            model_kwargs=kwargs,
        )

    def embed(self, texts: list[str]) -> list[list[float]]:
        vectors = self.model.encode(texts, normalize_embeddings=True, batch_size=2)
        return [row.tolist() for row in vectors]