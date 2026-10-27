from langgraph.graph import END, START, StateGraph

from shanks_agent.graph.nodes.writer import make_writer_node
from shanks_agent.graph.state import AgentState
from shanks_agent.llm.failover import FailoverLLM


def build(llm: FailoverLLM):
    g: StateGraph = StateGraph(AgentState)
    g.add_node("write", make_writer_node(llm))
    g.add_edge(START, "write")
    g.add_edge("write", END)
    return g.compile()
