package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/machine/client"
)

func processesHandler() server.ToolHandlerFunc {
	return func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		raw, err := json.MarshalIndent(client.ReadProcesses(), "", "  ")
		return textOrErr(string(raw), err)
	}
}
