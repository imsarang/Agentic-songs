from ecommerce.domain.models import ContentFingerprint, CrawlJob, FetchedPage, ParsedPage, StoredArtifact
from ecommerce.ports.store import ObjectStorePort


class StoreStage():
    def __init__(self, objectStorePort: ObjectStorePort):
        self._objectStorePort = objectStorePort

    def run(self, job: CrawlJob, fetched: FetchedPage, parsed: ParsedPage, fingerprint: ContentFingerprint) -> list[StoredArtifact]:
        pass