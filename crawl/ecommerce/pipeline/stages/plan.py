from ecommerce.domain.errors import InvalidJobError
from ecommerce.domain.models import CrawlJob, PagePlan


class PlanStage:
    def run(self, job: CrawlJob) -> list[PagePlan]:
        if not job.url or not job.url.strip():
            raise InvalidJobError(message="url is required for crawl", field="url")

        plan = PagePlan(
            url=job.url.strip(),
            host=job.host,
            origin=job.origin,
            category=job.category,
            job_type=job.type,
            kind=job.type,
            depth=0,
            parent_url=None,
        )
        return [plan]
