from shanks_agent.config import get_settings
from shanks_agent.graph.domains.common import build_domain_graph
from shanks_agent.graph.nodes.tools import STUDY_TOOLS
from shanks_agent.llm.failover import FailoverLLM
from shanks_agent.mcp.session import MCPSession

def build(llm: FailoverLLM):
    return build_domain_graph(
        llm, "study", STUDY_TOOLS, MCPSession(get_settings().retrieve_mcp_url),
    )