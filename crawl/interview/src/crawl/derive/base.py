from abc import ABC, abstractmethod

from crawl.models import CheckError, Derived, Fetched


class Parser(ABC):
    def derive(self, fetched: Fetched) -> Derived:
        text = self.text(fetched).strip()
        if not text:
            raise CheckError(
                "derived text is empty"
            )
        return Derived(
            body=text.encode(),
            content_type="text/plain",
            suffix="txt",
        )

    @abstractmethod
    def text(self, fetched: Fetched) -> str:
        pass