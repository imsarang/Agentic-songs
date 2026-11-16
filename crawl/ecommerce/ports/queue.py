from abc import ABC, abstractmethod

from ecommerce.domain.models import CrawlJob


class QueuePort(ABC):
    @abstractmethod
    def enqueue(self, job: CrawlJob) -> str:
        pass