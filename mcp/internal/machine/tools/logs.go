package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/machine/client"
	"shanks/internal/machine/constants"
)

func logsHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return jsonResult(client.ReadLogs(ctx, constants.LogQuery{
			Source:   req.GetString("source", "journal"),
			Since:    req.GetString("since", ""),
			Priority: req.GetString("priority", ""),
			Unit:     req.GetString("unit", ""),
			Match:    req.GetString("match", ""),
		}))
	}
}
