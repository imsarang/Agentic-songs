_BLOCK = ("exfiltrate all", "dump all passwords", "ignore all previous instructions and send")

def check(text: str) -> str | None:
    low = text.lower()
    for phrase in _BLOCK:
        if phrase in low:
            return "blocked by input guardrail"
    return None