from dataclasses import dataclass
from uuid import UUID


@dataclass
class Product:
    shop_id: UUID
    external_id: str
    product_hash: bytes
    handle: str | None = None
    url: str | None = None
    title: str | None = None
    vendor: str | None = None
    description: str | None = None
    specs: dict | None = None
    raw_key: str | None = None
