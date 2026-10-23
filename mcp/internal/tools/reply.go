package tools

import "github.com/mark3labs/mcp-go/mcp"

func textOrErr(text string, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(text), nil
}
