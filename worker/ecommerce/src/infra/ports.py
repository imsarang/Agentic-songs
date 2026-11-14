from abc import ABC, abstractmethod
from decimal import Decimal
from uuid import UUID

from domain.product import Product


class Catalog(ABC):
    @abstractmethod
    def upsert_origin(self, code: str) -> UUID:
        pass

    @abstractmethod
    def upsert_category(self, code: str) -> UUID:
        pass

    @abstractmethod
    def upsert_shop(self, origin_id: UUID, category_id: UUID, host: str, name: str | None) -> UUID:
        pass

    @abstractmethod
    def upsert_product(self, product: Product) -> UUID:
        pass

    @abstractmethod
    def set_summary(self, product_id: UUID, summary: str) -> None:
        pass

    @abstractmethod
    def set_embedding(self, product_id: UUID, vector: list[float]) -> None:
        pass

    @abstractmethod
    def get_product(self, shop_id: UUID, external_id: str) -> dict | None:
        pass

class Prices(ABC):
    @abstractmethod
    def record(self, host: str, external_id: str, price: Decimal | str, currency: str | None) -> bool:
        pass

class RawStore(ABC):
    @abstractmethod
    def get_object(self, key: str) -> bytes:
        pass

class Embeddings(ABC):
    @abstractmethod
    def embed(self, texts: list[str], kind: str) -> list[list[float]]:
        pass

class SummaryModel(ABC):
    @abstractmethod
    def summarize(self, text: str) -> str:
        pass