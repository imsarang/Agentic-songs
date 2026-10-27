from shanks_agent.middleware.auth import token as auth
from shanks_agent.middleware.gaurdrails import input as guard_in
from shanks_agent.middleware.ratelimit import token_bucket

def before_turn(*, text: str, thread_id: str, token: str | None) -> str | None:
    err = auth.check(token)
    if err:
        return err
    err = token_bucket.check(thread_id or "default")
    if err:
        return err
    return guard_in.check(text or "")