from datetime import datetime
from typing import Any

from pydantic import BaseModel


class CrawlJob(BaseModel):
    type: str
    origin: str
    category: str
    host: str
    url: str | None = None
    name: str | None = None
    key: str | None = None
    external_id: str | None = None
    handle: str | None = None
    title: str | None = None
    vendor: str | None = None
    description: str | None = None
    specs: dict[str, Any] | None = None
    price: str | int | float | None = None
    currency: str | None = None


class PagePlan(BaseModel):
    url: str
    host: str
    origin: str
    category: str
    job_type: str
    kind: str
    depth: int = 0
    parent_url: str | None = None


class FetchedPage(BaseModel):
    url: str
    final_url: str
    status_code: int
    content_type: str
    body: bytes
    fetched_at: datetime
    headers: dict[str, str] | None = None


class ParsedPage(BaseModel):
    url: str
    content_type: str
    text: str
    links: list[str]
    raw_format: str
    title: str | None = None


class ContentFingerprint(BaseModel):
    algorithm: str
    digest: str
    url: str


class DedupDecision(BaseModel):
    fingerprint: ContentFingerprint
    is_duplicate: bool
    prior_key: str | None = None


class StoredArtifact(BaseModel):
    bucket: str
    key: str
    kind: str
    content_type: str
    sha256: str
    byte_size: int


class CrawlResult(BaseModel):
    job: CrawlJob
    status: str
    url: str
    artifacts: list[StoredArtifact]
    started_at: datetime
    finished_at: datetime
    fingerprint: ContentFingerprint | None = None
    error_code: str | None = None
    error_message: str | None = None
