import json
import os
import urllib.request

from infra.ports import SummaryModel


class LocalSummarizer(SummaryModel):
    def summarize(self, text: str) -> str:
        url = os.environ["LOCAL_BASE_URL"].rstrip("/") + "/chat/completions"
        model = os.environ["LOCAL_MODEL"]
        key = os.environ.get("LOCAL_API_KEY", "local")
        raw = json.dumps({
            "model": model,
            "temperature": 0,
            "messages": [
                {"role": "system", "content": "Summarize this product in two or three sentences."},
                {"role": "user", "content": text},
            ],
        }).encode()
        request = urllib.request.Request(
            url,
            data=raw,
            headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {key}",
            },
        )
        with urllib.request.urlopen(request) as response:
            body = json.load(response)
        return body["choices"][0]["message"]["content"]