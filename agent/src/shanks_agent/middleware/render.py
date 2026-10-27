"""Turn a raw tool payload into a readable reply. No model call."""

from __future__ import annotations

import html
import json
import re
from dataclasses import dataclass, field

# Free-text values run to the end of the line, so a snippet can contain "k=".
_FREE = {"snippet", "body", "text", "description", "message", "error"}
_TITLE = ("subject", "title", "name")
_WHO = ("from", "artist", "email")
_WHEN = ("date",)
_BODY = ("snippet", "body", "text", "description", "message")
# Handles the user does not need to read. Shown only when nothing else is left.
_HIDE = {"id", "uri", "message_id", "draft_id", "search_id", "index"}

_START = re.compile(
    r"^(?:([A-Za-z][A-Za-z0-9_-]{0,31})\s+)?([A-Za-z_][A-Za-z0-9_]*)="
)
_NEXT = re.compile(r"\s([A-Za-z_][A-Za-z0-9_]*)=")

_SNIPPET_LIMIT = 220
_BODY_LIMIT = 2000
_TOTAL_LIMIT = 4000


@dataclass
class Rec:
    prefix: str
    fields: dict[str, str]
    extra: list[str] = field(default_factory=list)


def render(text: str) -> str:
    raw = (text or "").strip()
    if not raw:
        return ""
    data = _json(raw)
    if data is not None:
        out = _json_text(data)
    else:
        out = _lines(raw)
    out = out.strip()
    if not out:
        return raw
    if len(out) > _TOTAL_LIMIT:
        return out[:_TOTAL_LIMIT].rstrip() + "…"
    return out


def _json(raw: str):
    if not raw.startswith(("{", "[")):
        return None
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return None


def _lines(text: str) -> str:
    items: list[Rec | str] = []
    for line in text.splitlines():
        line = line.strip()
        if not line:
            continue
        parsed = _pairs(line)
        if parsed is None:
            if items and isinstance(items[-1], Rec):
                items[-1].extra.append(line)
            else:
                items.append(line)
            continue
        prefix, pairs = parsed
        items.append(Rec(prefix, dict(pairs)))

    blocks: list[str] = []
    i = 0
    while i < len(items):
        item = items[i]
        if isinstance(item, str):
            blocks.append(item)
            i += 1
            continue
        j = i + 1
        while j < len(items) and isinstance(items[j], Rec):
            j += 1
        blocks.append(_render_records(items[i:j]))
        i = j
    return "\n\n".join(block for block in blocks if block)


def _pairs(line: str) -> tuple[str, list[tuple[str, str]]] | None:
    start = _START.match(line)
    if not start:
        return None
    prefix = (start.group(1) or "").strip()
    pairs: list[tuple[str, str]] = []
    pos = start.start(2)
    while pos < len(line):
        key_match = re.match(r"([A-Za-z_][A-Za-z0-9_]*)=", line[pos:])
        if not key_match:
            return None
        key = key_match.group(1)
        val_at = pos + key_match.end()
        if key in _FREE:
            pairs.append((key, _clean(line[val_at:])))
            break
        nxt = _NEXT.search(line, val_at)
        if not nxt:
            pairs.append((key, _clean(line[val_at:])))
            break
        pairs.append((key, _clean(line[val_at:nxt.start()])))
        pos = nxt.start(1)
    return prefix, pairs


def _render_records(recs: list[Rec]) -> str:
    body = [rec for rec in recs if not _header_only(rec)]
    if not body:
        return ""

    as_list = len(body) > 1
    cards: list[str] = []
    for n, rec in enumerate(body, 1):
        number = rec.fields.get("index", str(n)) if as_list else None
        cards.append(_card(rec, number))
    listed = "\n\n".join(cards)
    if as_list and any("index" in rec.fields for rec in body):
        return f"{len(body)} results\n\n{listed}"
    return listed


def _header_only(rec: Rec) -> bool:
    # search_id is a handle for the next tool call, not something to read out.
    return not rec.prefix and not rec.extra and set(rec.fields) == {"search_id"}


def _card(rec: Rec, number: str | None) -> str:
    fields = rec.fields
    title_key = next((key for key in _TITLE if fields.get(key)), "")
    title = fields.get(title_key, "")
    who = _who(next((fields[key] for key in _WHO if fields.get(key)), ""))
    when = next((fields[key] for key in _WHEN if fields.get(key)), "")
    body_key = next((key for key in _BODY if fields.get(key)), "")
    body = fields.get(body_key, "")
    if body_key == "snippet":
        body = _clip(body, _SNIPPET_LIMIT)

    used = set(_HIDE)
    used.update(_TITLE)
    used.update(_WHO)
    used.update(_WHEN)
    used.update(_BODY)
    rest = [(key, value) for key, value in fields.items() if key not in used and value]

    rows: list[str] = []
    lone_title = bool(title) and not who and not when and not body and not rest and not rec.extra
    if lone_title:
        rows.append(f"{_label(title_key)}: {title}")
    elif title and rec.prefix:
        rows.append(f"{rec.prefix.capitalize()}: {title}")
    elif title:
        rows.append(title)
    elif rec.prefix:
        rows.append(rec.prefix.capitalize())

    meta = " · ".join(part for part in (who, when) if part)
    if meta and meta != (rows[0] if rows else ""):
        rows.append(meta)
    for key, value in rest:
        rows.append(f"{_label(key)}: {value}")
    if body:
        rows.append(body)
    extra = "\n".join(rec.extra).strip()
    if extra:
        rows.append(_clip(extra, _BODY_LIMIT))
    if not title and not who and not when and not body and not rest:
        for key, value in fields.items():
            if key in _HIDE and key not in {"index", "search_id"} and value:
                rows.append(f"{_label(key)}: {value}")

    if not rows:
        rows = [f"{_label(key)}: {value}" for key, value in fields.items() if value]
    if number is None:
        return "\n".join(rows)
    first, *tail = rows
    return "\n".join([f"{number}. {first}", *(f"   {line}" for line in tail)])


def _json_text(data: object, depth: int = 0) -> str:
    pad = "  " * depth
    if isinstance(data, dict):
        lines: list[str] = []
        for key, value in data.items():
            if value is None or value == "" or value == [] or value == {}:
                continue
            label = _label(str(key))
            if isinstance(value, (dict, list)):
                nested = _json_text(value, depth + 1)
                if nested:
                    lines.append(f"{pad}{label}")
                    lines.append(nested)
            else:
                lines.append(f"{pad}{label}: {_scalar(value)}")
        return "\n".join(lines)
    if isinstance(data, list):
        if not data:
            return ""
        if all(isinstance(item, dict) for item in data):
            blocks: list[str] = []
            for index, item in enumerate(data, 1):
                nested = _json_text(item, depth + 1)
                if not nested:
                    continue
                lines = nested.splitlines()
                head = f"{pad}{index}. {lines[0].strip()}"
                blocks.append("\n".join([head, *lines[1:]]))
            return "\n\n".join(blocks)
        return "\n".join(f"{pad}- {_scalar(item)}" for item in data)
    return f"{pad}{_scalar(data)}"


def _who(value: str) -> str:
    value = value.strip()
    if value.startswith("<") and value.endswith(">") and " " not in value:
        return value[1:-1]
    return value


def _clean(value: str) -> str:
    return html.unescape(" ".join(value.split()))


def _clip(text: str, limit: int) -> str:
    if len(text) <= limit:
        return text
    return text[:limit].rstrip() + "…"


def _label(key: str) -> str:
    return key.replace("_", " ").capitalize()


def _scalar(value: object) -> str:
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, float):
        return f"{value:g}"
    return _clean(str(value))
