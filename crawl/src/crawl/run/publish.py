from crawl.dedup.hash import sha256_hex
from crawl.derive import parser_for
from crawl.fetch.client import CrawlerClient
from crawl.fetch.plan import scrape_body
from crawl.models import CheckError, QueueMessage
from crawl.queue.client import QueueInterface
from crawl.store.keys import derived_key, raw_key
from crawl.store.s3 import S3Store


def publish_one(
    source_id: str,
    url: str,
    fetcher: CrawlerClient,
    store: S3Store,
    queue: QueueInterface,
) -> QueueMessage | None:


    if store.raw_for_url(source_id, url):
        return None

    fetched = fetcher.fetch(scrape_body(url))
    if not fetched.body:
        raise CheckError(f"{source_id} returned an empty body")

    derived = parser_for(fetched.content_type).derive(fetched)
    digest = sha256_hex(fetched.body)
    raw = raw_key(source_id, digest)
    derived_path = derived_key(source_id, digest, derived.suffix)
    meta = {"source_id": source_id, "url": url}
    store.put(raw, fetched.body, fetched.content_type, meta)
    store.put(derived_path, derived.body, derived.content_type, meta)
    message = QueueMessage(
        source_id=source_id,
        sha256=digest,
        key=derived_path,
        url=url,
    )
    queue.publish(message)
    return message
