package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func textOrErr(text string, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(text), nil
}

func jsonResult(v any, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return textOrErr("", err)
	}
	raw, err := json.MarshalIndent(v, "", "  ")
	return textOrErr(string(raw), err)
}
