RAW, SUMMARIZE, CONTINUE = "raw", "summarize", "continue"

def normalize(value: object) -> str:
    v = str(value or "").strip().lower()
    return v if v in (SUMMARIZE, CONTINUE) else RAW

def after_tool(state: dict) -> str:
    return "agent" if normalize(state.get("present")) in (SUMMARIZE, CONTINUE) else "end"