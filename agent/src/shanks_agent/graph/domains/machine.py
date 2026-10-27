from shanks_agent.config import get_settings
from shanks_agent.graph.domains.common import build_domain_graph
from shanks_agent.graph.nodes.tools import MACHINE_TOOLS
from shanks_agent.llm.failover import FailoverLLM
from shanks_agent.mcp.session import MCPSession

def build(llm: FailoverLLM):
    return build_domain_graph(
        llm, "machine", MACHINE_TOOLS, MCPSession(get_settings().machine_mcp_url),
    )