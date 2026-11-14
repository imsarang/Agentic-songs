import hashlib
import json

from domain.job import Job
from domain.product import Product
from infra.ports import Catalog, Embeddings, Prices, RawStore


class CrawlJob:
    def __init__(self, catalog: Catalog, prices: Prices, raw: RawStore, embeddings: Embeddings) -> None:
        self._catalog = catalog
        self._prices = prices
        self._raw = raw
        self._embeddings = embeddings

    def run(self, job: Job) -> None:
        if job.key is None:
            raise ValueError("key")
        body = self._raw.get_object(job.key)
        digest = hashlib.sha256(body).digest()
        payload = json.loads(body)
        shop_id = self._shop(job)
        existing = self._catalog.get_product(shop_id, payload["external_id"])
        if existing is not None and bytes(existing["product_hash"]) == digest:
            return
        product_id = self._catalog.upsert_product(
            Product(
                shop_id=shop_id,
                external_id=payload["external_id"],
                product_hash=digest,
                handle=payload.get("handle"),
                url=job.url,
                title=payload.get("title"),
                vendor=payload.get("vendor"),
                description=payload.get("description"),
                specs=payload.get("specs"),
                raw_key=job.key,
            )
        )
        vector = self._embeddings.embed([self._text(payload)], "document")[0]
        self._catalog.set_embedding(product_id, vector)
        if payload.get("price") is not None:
            self._prices.record(
                job.host,
                payload["external_id"],
                payload["price"],
                payload.get("currency"),
            )

    def _shop(self, job: Job):
        origin_id = self._catalog.upsert_origin(job.origin)
        category_id = self._catalog.upsert_category(job.category)
        return self._catalog.upsert_shop(origin_id, category_id, job.host, job.name)

    def _text(self, payload: dict) -> str:
        return "\n".join(
            part
            for part in (
                payload.get("title") or "",
                payload.get("vendor") or "",
                payload.get("description") or "",
                json.dumps(payload.get("specs") or {}, sort_keys=True),
            )
            if part
        )