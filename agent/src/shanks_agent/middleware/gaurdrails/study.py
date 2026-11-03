import json

EMPTY = "I found nothing in the study pages for that question."
FAILED = "Retrieval failed. I can't answer from the study pages."
UNGROUNDED = "I can't answer that from the retrieved pages."


def _results(messages) -> list[dict]:
    found = []
    for message in messages or []:
        content = getattr(message, "content", "")
        if not isinstance(content, str) or "[tool retrieve result]" not in content:
            continue
        raw = content.split("\n", 1)[-1]
        try:
            data = json.loads(raw)
        except json.JSONDecodeError:
            return [{"_bad": True}]
        found.extend(data.get("chunks") or [])
    return found


def closed(messages) -> str | None:
    if not any(
        isinstance(getattr(m, "content", ""), str) and "[tool retrieve result]" in m.content
        for m in messages or []
    ):
        return None
    chunks = _results(messages)
    if chunks and chunks[0].get("_bad"):
        return FAILED
    if not chunks:
        return EMPTY
    return None


def ground(text: str, messages) -> str:
    urls = {c.get("url") for c in _results(messages) if c.get("url")}
    if urls and any(url in text for url in urls):
        return text
    return UNGROUNDED