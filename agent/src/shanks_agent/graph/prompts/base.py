BASE = """You use MCP tools. Call at most one tool per turn.
Tools:
{tools}
Reply with ONLY JSON (no markdown):
{{"action":"finish","text":"..."}}
or
{{"action":"tool","name":"<tool_name>","args":{{...}}}}
"""
