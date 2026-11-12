import os
from pathlib import Path

from langchain_openai import ChatOpenAI

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "synthetic.json"
GRAPH = ROOT /"graph.json"

EMBED_URL = os.environ.get("EMBED_URL", "http://127.0.0.1:8091/embed")
AGENT_URL = os.environ.get("AGENT_URL", "http://127.0.0.1:8001")
S3_ENDPOINT = os.environ.get("S3_ENDPOINT", "http://127.0.0.1:9000")
S3_BUCKET = os.environ.get("S3_BUCKET", "shanks-raw")
S3_ACCESS_KEY = os.environ.get("S3_ACCESS_KEY", "shanks")
S3_SECRET_KEY = os.environ.get("S3_SECRET_KEY", "shanks-secret")
S3_REGION = os.environ.get("S3_REGION", "us-east-1")
LM_BASE = os.environ.get("LOCAL_BASE_URL", "http://127.0.0.1:1234/v1")
LM_MODEL = os.environ.get("LOCAL_MODEL", "qwen2.5-7b-instruct")


def model() -> ChatOpenAI:
    return ChatOpenAI(
        base_url=LM_BASE,
        api_key="lm-studio",
        model=LM_MODEL,
        temperature=0,
    )
