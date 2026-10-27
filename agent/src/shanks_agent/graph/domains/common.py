from langgraph.graph import END, START, StateGraph

from shanks_agent.graph.nodes.agent import make_agent_node
from shanks_agent.graph.nodes.tools import make_tool_node
from shanks_agent.graph.state import AgentState
from shanks_agent.llm.failover import FailoverLLM
from shanks_agent.mcp.session import MCPSession

def build_domain_graph(llm: FailoverLLM, domain: str, tools: tuple[str, ...], mcp: MCPSession):
    tool_nodes = {f"tool_{n}": make_tool_node(n, mcp) for n in tools}
    allowed = set(tools)

    def route(state: AgentState) -> str:
        name = state.get("next_tool")
        if name and name in allowed:
            return f"tool_{name}"
        return END

    g: StateGraph = StateGraph(AgentState)
    g.add_node("agent", make_agent_node(llm))
    for nid, fn in tool_nodes.items():
        g.add_node(nid, fn)
    g.add_edge(START, "agent")
    g.add_conditional_edges("agent", route, {**{nid: nid for nid in tool_nodes}, END: END})
    for nid in tool_nodes:
        g.add_edge(nid, "agent")
    return g.compile()