def scrape_body(url: str) -> dict:
    return {"url": url, "formats": ["rawBase64"], "parsers": []}