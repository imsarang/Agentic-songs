import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer

import torch
from sentence_transformers import SentenceTransformer

QUERY = "Represent this sentence for searching relevant passages: "
device = "cuda" if torch.cuda.is_available() else "cpu"
kwargs = {"torch_dtype": torch.float16} if device == "cuda" else {}
model = SentenceTransformer(os.environ["EMBED_MODEL_PATH"], device=device, model_kwargs=kwargs)


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        size = int(self.headers["Content-Length"])
        body = json.loads(self.rfile.read(size))
        texts = body["texts"]
        if body.get("kind") == "query":
            texts = [QUERY + text for text in texts]
        vectors = model.encode(texts, normalize_embeddings=True, batch_size=2)
        raw = json.dumps({"vectors": [row.tolist() for row in vectors]}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)


if __name__ == "__main__":
    print(f"embed on {device}")
    HTTPServer(("127.0.0.1", 8091), Handler).serve_forever()