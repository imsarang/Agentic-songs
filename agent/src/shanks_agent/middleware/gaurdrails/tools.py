from shanks_agent.middleware.policy.domain_defaults import letter_to_send

# tools that must not run unless user clearly asked this turn
_CONFIRM = {"google_send_mail", "google_delete_mail", "google_delete_drafts"}

def check(name: str, user_text: str) -> str | None:
    if name not in _CONFIRM:
        return None
    low = user_text.lower()
    if name == "google_send_mail" and not letter_to_send(user_text) and not any(
        w in low for w in ("send", "yes", "do it", "go ahead")
    ):
        return "send needs confirm"
    if "delete" in name and "delete" not in low and "trash" not in low and "yes" not in low:
        return "delete needs confirm"
    return None