from abc import ABC, abstractmethod

from crawl.models import QueueMessage


class QueueInterface(ABC):
    @abstractmethod
    def publish(self, message: QueueMessage) -> None:
        pass
