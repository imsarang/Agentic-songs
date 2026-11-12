from crawl.config import Settings
from crawl.fetch.client import CrawlerClient
from crawl.models import CheckError, Fetched

import httpx
import base64

class FireCrawl(CrawlerClient):
    def __init__(self, settings: Settings) -> None:
        self.settings = settings

    def fetch(self, body: dict) -> Fetched:
        response = httpx.post(
            f"{self.settings.firecrawl_api_url.rstrip('/')}/v2/scrape",
            headers={
                "Authorization": f"Bearer {self.settings.firecrawl_api_key}"
            },
            json = body,
            timeout = 180,
        )

        response.raise_for_status()

        payload = response.json()
        if payload.get("success") is False:
            raise CheckError(str(payload.get("error") or "Scrape failed"))
        data = payload["data"]

        encoded = data.get("rawBase64")

        if not encoded:
            raise CheckError("scrape returned no rawBase64")
        content_type = str((data.get("metadata") or {}).get("contentType") or "application/octet-stream")
        return Fetched(
            body=base64.b64decode(encoded),
            content_type=content_type.split(";", 1)[0].strip()
        )
