from io import BytesIO
from crawl.derive.base import Parser
from crawl.models import Fetched

from pypdf import PdfReader

class PdfParser(Parser):
    def text(self, fetched: Fetched) -> str:
        reader = PdfReader(BytesIO(fetched.body))
        return "\n".join(page.extract_text() or "" for page in reader.pages)
