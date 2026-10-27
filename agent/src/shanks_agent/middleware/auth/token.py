import os

def check(token: str | None) -> str | None:
    expected = os.getenv("AGENT_TURN_TOKEN", "").strip()
    if not expected: return None
    if (token or "").strip() != expected:
        return "unauthorized"
    return None