def raw_key(source_id: str, digest: str) -> str:
    return f"raw/{source_id}/{digest}"


def derived_key(source_id: str, digest: str, suffix: str) -> str:
    return f"text/{source_id}/{digest}.{suffix}"