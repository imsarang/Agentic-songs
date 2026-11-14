from domain.job import Job
from domain.product import Product
from infra.ports import Catalog, Prices, SummaryModel

import json
import hashlib

class ApiJob:
    def __init__(self, catalog: Catalog, prices: Prices, summaries: SummaryModel) -> None:
        self._catalog = catalog
        self._prices = prices
        self._summaries = summaries

    def run(self, job: Job) -> None:
        if job.external_id is None:
            raise ValueError("external_id")
        shop_id = self._shop(job)
        product = Product(
            shop_id=shop_id,
            external_id=job.external_id,
            product_hash=self._hash(job),
            handle=job.handle,
            url=job.url,
            title=job.title,
            vendor=job.vendor,
            description=job.description,
            specs=job.specs,
        )
        product_id = self._catalog.upsert_product(product)
        if job.price is not None:
            self._prices.record(job.host, job.external_id, job.price, job.currency)
        self._catalog.set_summary(product_id, self._summaries.summarize(self._text(job)))

    def _shop(self, job: Job):
        origin_id = self._catalog.upsert_origin(job.origin)
        category_id = self._catalog.upsert_category(job.category)
        return self._catalog.upsert_shop(origin_id, category_id, job.host, job.name)

    def _hash(self, job: Job) -> bytes:
        raw = json.dumps(
            {
                "host": job.host,
                "external_id": job.external_id,
                "title": job.title,
                "vendor": job.vendor,
                "description": job.description,
                "specs": job.specs,
                "price": job.price,
            },
            sort_keys=True,
        ).encode()
        return hashlib.sha256(raw).digest()

    def _text(self, job: Job) -> str:
        return "\n".join(
            part
            for part in (
                job.title or "",
                job.vendor or "",
                job.description or "",
                json.dumps(job.specs or {}, sort_keys=True),
            )
            if part
        )