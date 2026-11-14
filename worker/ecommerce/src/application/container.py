from dataclasses import dataclass

from infra.catalog import PostgresCatalog
from infra.embed import Embedder
from infra.llm import LocalSummarizer
from infra.ports import Catalog, Embeddings, Prices, RawStore, SummaryModel
from infra.price_history import PriceHistory
from infra.s3 import ObjectStore


@dataclass
class Clients:
    catalog: Catalog
    prices: Prices
    raw: RawStore
    embeddings: Embeddings
    summaries: SummaryModel

def build() -> Clients:
    return Clients(
        catalog=PostgresCatalog(),
        prices=PriceHistory(),
        raw=ObjectStore(),
        embeddings=Embedder(),
        summaries=LocalSummarizer()
    )