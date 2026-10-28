from crawl.models import Source


def scrape_body(source: Source) -> dict:
    return {"url": str(source.url), "formats": ["rawBase64"], "parsers": []}