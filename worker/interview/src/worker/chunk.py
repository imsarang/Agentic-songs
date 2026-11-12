import os

from transformers import AutoTokenizer


class Chunker:
    def __init__(self, size: int = 510, overlap: int = 50) -> None:
        self.size = size
        self.overlap = overlap
        self.tokenizer = AutoTokenizer.from_pretrained(os.environ["EMBED_MODEL_PATH"])

    def split(self, text: str) -> list[str]:
        parts = []
        for section in text.split("<<section>>"):
            section = section.strip()
            if not section:
                continue

            ids = self.tokenizer.encode(section, add_special_tokens=False)
            
            # if size of chunk is less than max chunk size
            if len(ids) <= self.size:
                parts.append(self.tokenizer.decode(ids).strip())
                continue

            step = self.size - self.overlap
        
            for start in range(0, len(ids), step):
                piece = self.tokenizer.decode(ids[start:start + self.size]).strip()
                if piece:
                    parts.append(piece)
                if start + self.size >= len(ids):
                    break
        return parts
