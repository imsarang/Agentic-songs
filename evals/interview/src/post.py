import json
import urllib.request


def open_json(url: str, payload: dict):
    body = json.dumps(payload).encode()
    request = urllib.request.Request(
        url,
        data=body,
        headers={"Content-Type": "application/json"},
    )
    return urllib.request.urlopen(request)
