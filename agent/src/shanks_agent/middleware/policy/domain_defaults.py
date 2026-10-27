_WRITE = ("write", "compose", "draft")
_PIECE = ("letter", "mail", "email", "e-mail")

def letter_to_send(user_text: str) -> bool:
    low = user_text.lower()
    if low.startswith(("recent ", "creative draft", "letter flow")):
        return False
    return any(w in low for w in _WRITE) and any(w in low for w in _PIECE)