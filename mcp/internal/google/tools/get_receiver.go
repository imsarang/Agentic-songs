package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
)

func getReceiverHandler(client google.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		rows, err := client.GetReciever(ctx, req.GetString("name", ""))
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr(google.FormatContacts(rows), nil)
	}
}
