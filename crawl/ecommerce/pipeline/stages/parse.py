from ecommerce.domain.models import FetchedPage, ParsedPage


class ParseStage():
    def __init__(self):
        pass

    def run(self, page: FetchedPage) -> ParsedPage:
        pass
    