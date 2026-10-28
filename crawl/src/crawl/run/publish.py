import json
from io import BytesIO

from pypdf import PdfReader

from crawl.dedup.hash import sha256_hex
from crawl.fetch.client import CrawlerClient
from crawl.fetch.plan import scrape_body
from crawl.models import CheckError, Derived, Fetched, QueueMessage, Source, StatementRow
from crawl.queue.client import QueueInterface
from crawl.store.keys import derived_key, raw_key
from crawl.store.s3 import S3Store


def publish_one(
    source: Source,
    fetcher: CrawlerClient,
    store: S3Store,
    queue: QueueInterface
) -> QueueMessage:
    fetched = fetcher.fetch(scrape_body(source))
    if not fetched.body:
        raise CheckError(f"{source.id} returned an empty body")

    derived = divide(source, fetched)
    digest = sha256_hex(fetched.body)
    raw = raw_key(source.id, digest)
    derived_path = derived_key(
        source.kind,
        source.id,
        digest,
        derived.suffix
    )

    meta = {
        "source_id": source.id,
        "symbol": source.symbol,
        "fiscal_year": str(source.fiscal_year),
        "period": source.period,
        "scope": source.scope,
        "doc_type": source.doc_type,
    }

    store.put(raw, fetched.body, fetched.content_type, meta)
    store.put(derived_path, derived.body, derived.content_type, meta)
    message = QueueMessage(
        source_id=source.id,
        sha256=digest,
        key=derived_path,
        kind=source.kind,
        symbol=source.symbol,
        fiscal_year=source.fiscal_year,
        period=source.period,
        scope=source.scope,
        doc_type=source.doc_type,
        url=str(source.url),
    )
    queue.publish(message)
    return message


def divide(source: Source, fetched: Fetched) -> Derived:
    if source.kind == "narrative":
        text = narrative_text(fetched).strip()
        if not text:
            raise CheckError(f"{source.id} narrative is empty")
        return Derived(kind="narrative", body=text.encode(), content_type="text/plain", suffix="txt")
    rows = statement_rows(fetched)
    if not rows:
        raise CheckError(f"{source.id} statement has no rows")
    body = json.dumps([row.model_dump() for row in rows]).encode()
    return Derived(kind="statement", body=body, content_type="application/json", suffix="json")

def narrative_text(fetched: Fetched) -> str:
    if fetched.content_type == "application/pdf":
        reader = PdfReader(BytesIO(fetched.body))
        return "\n".join(page.extract_text() or "" for page in reader.pages)
    if fetched.content_type in {"text/plain", "text/html"}:
        return fetched.body.decode("utf-8")
    raise CheckError(f"narrative source returned {fetched.content_type}")

def statement_rows(fetched: Fetched) -> list[StatementRow]:
    if fetched.content_type != "application/json":
        raise CheckError(f"statement source returned {fetched.content_type}")
    payload = json.loads(fetched.body)
    if not isinstance(payload, list):
        raise CheckError("statement body must be a list of rows")
    return [StatementRow.model_validate(row) for row in payload]