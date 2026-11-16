from abc import ABC, abstractmethod

from ecommerce.domain.models import StoredArtifact


class ObjectStorePort(ABC):
    @abstractmethod
    def put(self, key: str, body: bytes, content_type: str, kind: str) -> StoredArtifact:
        pass

    @abstractmethod
    def get(self, key: str) -> bytes:
        pass

    @abstractmethod
    def exists(self, key: str) -> bool:
        pass