from crawl.derive.base import Parser
from crawl.derive.html import HtmlParser
from crawl.derive.pdf import PdfParser
from crawl.derive.plain import PlainParser
from crawl.models import CheckError

_PARSERS: dict[str, type[Parser]] = {
    "application/pdf": PdfParser,
    "text/html": HtmlParser,
    "text/plain": PlainParser,
}


def parser_for(content_type: str) -> Parser:
    kind = _PARSERS.get(content_type)
    if kind is None:
        raise CheckError(f"no parser for {content_type}")
    return kind()