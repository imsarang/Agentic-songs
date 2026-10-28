
from abc import ABC, abstractmethod

from crawl.models import Fetched


class CrawlerClient(ABC):
    @abstractmethod
    def fetch(self, body: dict) -> Fetched:
        pass