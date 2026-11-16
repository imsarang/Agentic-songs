from ecommerce.domain.models import FetchedPage, PagePlan
from ecommerce.ports.fetch import FetchPort


class FetchStage():
    def __init__(self, fetchPort: FetchPort):
        self._fetchPort = fetchPort

    def run(plan: PagePlan) -> FetchedPage:
        pass