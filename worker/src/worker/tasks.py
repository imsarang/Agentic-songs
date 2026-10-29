from celery.signals import worker_process_init
from celery.utils.log import get_task_logger

from worker.app import celery
from worker.chunk import Chunker
from worker.embed import Embedder
from worker.objects import get_object
from worker.vectors import VectorStore

log = get_task_logger(__name__)

chunker = Chunker()
store = VectorStore()
embedder: Embedder | None = None

@worker_process_init.connect
def _boot(** _kwargs) -> None:
    global embedder
    embedder = Embedder()
    log.info("model ready on %s", embedder.device)

@celery.task(name="documents.chunk")
def store_document(message: dict) -> int:
    log.info("job %s %s", message["source_id"], message["key"])
    text = get_object(message["key"]).decode()
    parts = chunker.split(text)
    log.info("chunks %s chars %s", len(parts), len(text))
    stored = store.upsert(message, parts, embedder.embed(parts))
    log.info("stored %s chunks in study", stored)
    return stored