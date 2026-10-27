BASE = """You use MCP tools. Call at most one tool per turn.
Tools:
{tools}
Reply with ONLY JSON (no markdown):
{{"action":"finish","text":"..."}}
or
{{"action":"tool","name":"<tool_name>","args":{{...}}}}

Each tool result comes back to you. The user does not see it.
Call another tool when you still need one, then finish with the reply they should read.
"""