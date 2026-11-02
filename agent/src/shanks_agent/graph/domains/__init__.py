from shanks_agent.graph.domains import drive, google, machine, spotify, study
# from shanks_agent.graph.domains import creative
from shanks_agent.llm.failover import FailoverLLM

def build_all(llm: FailoverLLM) -> dict:
    return {
        "spotify": spotify.build(llm),
        "machine": machine.build(llm),
        "google": google.build(llm),
        "drive": drive.build(llm),
        # "creative": creative.build(llm),
        "study": study.build(llm)
    }