from abc import ABC, abstractmethod

from ecommerce.domain.models import FetchedPage


class FetchPort(ABC):
    @abstractmethod
    def fetch(self, url: str) -> FetchedPage:
        pass
