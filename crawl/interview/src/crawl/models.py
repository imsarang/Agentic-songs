from pydantic import BaseModel


class Fetched(BaseModel):
    body: bytes
    content_type: str


class Derived(BaseModel):
    body: bytes
    content_type: str
    suffix: str


class QueueMessage(BaseModel):
    source_id: str
    sha256: str
    key: str
    url: str


class CheckError(Exception):
    pass
