from crawl.derive.base import Parser
from crawl.models import Fetched

from html.parser import HTMLParser

_BLOCKS = {"p", "div", "li", "br", "h1", "h2", "h3", "h4", "h5", "h6"}
_HEADINGS = {"h1", "h2", "h3", "h4", "h5", "h6"}
_SKIP = {"script", "style"}

class _Page(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.parts: list[str] = []
        self.skip = 0

    def handle_starttag(self, tag: str, attrs) -> None:
        if tag in _SKIP:
            self.skip += 1
        if tag in _BLOCKS:
            self.parts.append("\n")
    
    def handle_endtag(self, tag: str) -> None:
        if tag in _SKIP and self.skip:
            self.skip -= 1
        if tag in _HEADINGS:
            self.parts.append("\n")
    
    def handle_data(self, data: str) -> None:
        if not self.skip:
            self.parts.append(data)

class HtmlParser(Parser):
    def text(self, fetched: Fetched) -> str:
        page = _Page()
        page.feed(fetched.body.decode("utf-8", errors="replace"))
        return "".join(page.parts)