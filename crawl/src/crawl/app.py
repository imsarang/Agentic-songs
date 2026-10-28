from crawl.config import Settings, load_sources
from crawl.fetch.firecrawl import FireCrawl
from crawl.queue.broker import CeleryQueue
from crawl.run.publish import publish_one
from crawl.store.s3 import S3Store


def main() -> None:
    settings = Settings()
    sources = load_sources(settings.sources_path)
    store = S3Store(settings)
    queue = CeleryQueue(settings)
    store.ensure_bucket()
    fetcher = FireCrawl(settings)
    for source in sources:
        message = publish_one(source, fetcher, store, queue)
        print(message.model_dump_json())
