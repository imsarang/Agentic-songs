def raw_key(source_id: str, digest: str) -> str:
    return f"raw/{source_id}/{digest}"


def derived_key(kind: str, source_id: str, digest: str, suffix: str) -> str:
    return f"{kind}/{source_id}/{digest}.{suffix}"