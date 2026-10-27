from typing import Annotated, Any, TypedDict
from langchain_core.messages import BaseMessage
from langgraph.graph.message import add_messages

class AgentState(TypedDict, total=False):
    messages: Annotated[list[BaseMessage], add_messages]
    domain: str  # spotify | machine | google | creative
    next_tool: str | None
    tool_args: dict[str, Any]
    present: str