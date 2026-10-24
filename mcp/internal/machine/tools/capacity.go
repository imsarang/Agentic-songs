package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/machine/client"
)

func capacityHandler() server.ToolHandlerFunc {
	return func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spec, err := client.ReadCapacity()
		if err != nil {
			return textOrErr("", err)
		}
		raw, err := json.MarshalIndent(spec, "", "  ")
		return textOrErr(string(raw), err)
	}
}
