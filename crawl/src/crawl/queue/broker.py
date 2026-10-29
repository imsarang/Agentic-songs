from celery import Celery

from crawl.config import Settings
from crawl.models import QueueMessage
from crawl.queue.client import QueueInterface

_TASK = "documents.chunk"


class CeleryQueue(QueueInterface):
    def __init__(self, settings: Settings) -> None:
        self._celery = Celery("crawl", broker=settings.celery_broker_url)

    def publish(self, message: QueueMessage) -> None:
        self._celery.send_task(_TASK, args=[message.model_dump(mode="json")])
