import logging

import boto3
import tiktoken
from langchain_core.documents import Document

from config import S3_ACCESS_KEY, S3_BUCKET, S3_ENDPOINT, S3_REGION, S3_SECRET_KEY

log = logging.getLogger(__name__)
LIMIT = 2000
MIN = 500
_ENC = tiktoken.get_encoding("cl100k_base")


def _pieces(text: str) -> list[str]:
    ids = _ENC.encode(text)
    pieces = [
        _ENC.decode(ids[start:start + LIMIT])
        for start in range(0, len(ids), LIMIT)
    ]
    if len(pieces) > 1 and len(_ENC.encode(pieces[-1])) < MIN:
        tail = len(_ENC.encode(pieces[-1]))
        if len(_ENC.encode(pieces[-2])) + tail <= LIMIT:
            log.info("join short tail tokens=%s", tail)
            pieces[-2] += pieces[-1]
            pieces.pop()
    return pieces

def pages() -> list[Document]:
    client = boto3.client(
        "s3",
        endpoint_url=S3_ENDPOINT,
        aws_access_key_id=S3_ACCESS_KEY,
        aws_secret_access_key=S3_SECRET_KEY,
        region_name=S3_REGION,
    )
    found = client.list_objects_v2(Bucket=S3_BUCKET, Prefix="text/")
    docs = []
    for item in found.get("Contents") or []:
        key = item["Key"]
        if not key.endswith(".txt"):
            continue
        text = client.get_object(Bucket=S3_BUCKET, Key=key)["Body"].read().decode()
        
        for part, piece in enumerate(_pieces(text)):
            log.info("page %s part %s tokens %s", key, part, len(_ENC.encode(piece)))
            docs.append(Document(page_content=piece, metadata={"key": key, "part": part}))

    if not docs:
        raise SystemExit("no derived text in minio")
    return docs
