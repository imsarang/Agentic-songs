from dataclasses import dataclass
from uuid import UUID

from psycopg.types.json import Json

from domain.product import Product
from infra.db import catalog_db
from infra.ports import Catalog


@dataclass
class PostgresCatalog(Catalog):
    def __init__(self) -> None:
        self._db = catalog_db

    def upsert_origin(self, code: str) -> UUID:
        row = self._db.run(
            """
            insert into origin (code) values (%s)
            on conflict (code) do update set code = excluded.code
            returning id
            """,
            (code,),
            one=True,
        )
        return row["id"]

    def upsert_category(self, code: str) -> UUID:
        row = self._db.run(
            """
            insert into category (code) values (%s)
            on conflict (code) do update set code = excluded.code
            returning id
            """,
            (code,),
            one=True
        )

        return row["id"]

    def upsert_product(self, product: Product) -> UUID:
        row = self._db.run(
            """
            insert into product (
            shop_id, external_id, handle, url, title, vendor, description,
            specs, raw_key, product_hash
            )
            values (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            on conflict (shop_id, external_id) do update
            set handle = excluded.handle,
                url = excluded.url,
                title = excluded.title,
                vendor = excluded.vendor,
                description = excluded.description,
                specs = excluded.specs,
                raw_key = excluded.raw_key,
                product_hash = excluded.product_hash,
                updated_at = now()
            returning id
            """,
            (
                product.shop_id,
                product.external_id,
                product.handle,
                product.url,
                product.title,
                product.vendor,
                product.description,
                Json(product.specs) if product.specs is not None else None,
                product.raw_key,
                product.product_hash,
            ),
            one=True,
        )

        return row["id"]

    def upsert_shop(self, origin_id: UUID, category_id: UUID, host: str, name: str | None) -> UUID:
        row = self._db.run(
            """
            insert into shop (origin_id, category_id, host, name)
            values (%s, %s, %s, %s)
            on conflict (host) do update
            set origin_id = excluded.origin_id,
                category_id = excluded.category_id,
                name = excluded.name
            returning id
            """,
            (origin_id, category_id, host, name),
            one=True,
        )
        return row["id"]

    def get_product(self, shop_id: UUID, external_id: str) -> dict | None:
        row = self._db.run(
            """
            select * from product
            where shop_id = %s and external_id = %s
            """,
            (
                shop_id,
                external_id
            ),
            one=True
        )

        return row

    def set_summary(self, product_id: UUID, summary: str) -> None:
        self._db.run(
            """
            update product set summary = %s, updated_at = now() where id = %s
            """,
            (summary, product_id)
        )
    
    def set_embedding(self, product_id: UUID, vector: list[float]) -> None:
        if len(vector) != 1024:
            raise ValueError(f"Embedding length {len(vector)}")

        literal = "[" + ",".join(str(item) for item in vector) + "]"
        self._db.run(
            """
            update product
            set embedding = %s::vector, updated_at = now()
            where id = %s
            """,
            (literal, product_id),
        )