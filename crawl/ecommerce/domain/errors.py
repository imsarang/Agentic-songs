class DomainError(Exception):
    def __init__(self, message: str) -> None:
        self.message = message
        super().__init__(message)


class InvalidJobError(DomainError):
    def __init__(self, message: str, field: str | None = None) -> None:
        self.field = field
        super().__init__(message)


class FetchError(DomainError):
    def __init__(
        self,
        message: str,
        url: str,
        status_code: int | None = None,
    ) -> None:
        self.url = url
        self.status_code = status_code
        super().__init__(message)


class ParseError(DomainError):
    def __init__(
        self,
        message: str,
        url: str,
        content_type: str | None = None,
    ) -> None:
        self.url = url
        self.content_type = content_type
        super().__init__(message)


class DedupError(DomainError):
    pass


class StoreError(DomainError):
    def __init__(self, message: str, key: str | None = None) -> None:
        self.key = key
        super().__init__(message)


class PipelineError(DomainError):
    def __init__(self, message: str, stage: str) -> None:
        self.stage = stage
        super().__init__(message)
