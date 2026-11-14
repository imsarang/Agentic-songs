from celery.utils.log import get_task_logger
from application.app import celery
from application.container import build
from domain.job import Job
from usecase.api import ApiJob
from usecase.crawl import CrawlJob

log = get_task_logger(__name__)
clients = build()

@celery.task(bind=True, name="ecommerce.upsert", max_retries=3)
def upsert(self, message: dict) -> None:
    try:
        job = Job.from_dict(message)
        log.info("job %s %s", job.type, job.external_id or job.key)
        if job.type == "api":
            ApiJob(clients.catalog, clients.prices, clients.summaries).run(job)
            return
        if job.type == "crawl":
            CrawlJob(clients.catalog, clients.prices, clients.raw, clients.embeddings).run(job)
            return
        raise ValueError(job.type)
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            raise
        raise self.retry(exc=exc, countdown=2**self.request.retries)