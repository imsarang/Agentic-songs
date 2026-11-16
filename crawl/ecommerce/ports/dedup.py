from abc import ABC, abstractmethod

from ecommerce.domain.models import ContentFingerprint, DedupDecision


class DedupPort(ABC):
    @abstractmethod
    def seen(self, fingerprint: ContentFingerprint) -> DedupDecision:
        pass

    @abstractmethod
    def mark(self, fingerprint: ContentFingerprint, key: str | None = None) -> None:
        pass