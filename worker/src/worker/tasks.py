from celery.utils.log import get_task_logger

from worker.app import celery
from worker.chunk import Chunker
from worker.embed import embed
from worker.objects import get_object
from worker.vectors import VectorStore

log = get_task_logger(__name__)

chunker = Chunker()
store = VectorStore()


@celery.task(name="documents.chunk.dead")
def dead_document(message: dict, error: str) -> None:
    log.error("dead %s %s %s", message.get("source_id"), message.get("sha256"), error)


@celery.task(bind=True, name="documents.chunk", max_retries=3)
def store_document(self, message: dict) -> int:
    try:
        if store.has_sha256(message["sha256"]):
            log.info("skip %s", message["sha256"])
            return 0
        log.info("job %s %s", message["source_id"], message["key"])
        text = get_object(message["key"]).decode()
        parts = chunker.split(text)
        log.info("chunks %s chars %s", len(parts), len(text))
        stored = store.upsert(message, parts, embed(parts, "document"))
        log.info("stored %s chunks in study", stored)
        return stored
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            dead_document.delay(message, str(exc))
            raise
        raise self.retry(exc=exc, countdown=2 ** self.request.retries)
