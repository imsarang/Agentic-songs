from shanks_agent.graph.domains import creative, drive, google, machine, spotify
from shanks_agent.llm.failover import FailoverLLM

def build_all(llm: FailoverLLM) -> dict:
    return {
        "spotify": spotify.build(llm),
        "machine": machine.build(llm),
        "google": google.build(llm),
        "drive": drive.build(llm),
        "creative": creative.build(llm),
    }