import os

from transformers import AutoTokenizer


class Chunker:
    def __init__(self, size: int = 510, overlap: int = 50) -> None:
        self.size = size
        self.overlap = overlap
        self.tokenizer = AutoTokenizer.from_pretrained(os.environ["EMBED_MODEL_PATH"])

    def split(self, text: str) -> list[str]:
        ids = self.tokenizer.encode(text, add_special_tokens=False)
        step = self.size - self.overlap
        parts = []
        for start in range(0, len(ids), step):
            piece = self.tokenizer.decode(ids[start:start + self.size]).strip()
            if piece:
                parts.append(piece)
            if start + self.size >= len(ids):
                break
        return parts
