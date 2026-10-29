from crawl.derive.base import Parser
from crawl.models import Fetched


class PlainParser(Parser):
    def text(self, fetched: Fetched) -> str:
        return fetched.body.decode("utf-8")