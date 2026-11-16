from ecommerce.domain.models import ContentFingerprint, DedupDecision, FetchedPage
from ecommerce.ports.dedup import DedupPort


class DedupStage():
    def __init__(self, dedupPort: DedupPort):
        self._dedupPort = dedupPort
    
    def fingerprint(self, page: FetchedPage) -> ContentFingerprint:
        pass

    def run(self, fingerprint: ContentFingerprint) -> DedupDecision:
        pass