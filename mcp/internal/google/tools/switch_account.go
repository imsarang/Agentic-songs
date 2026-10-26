package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
)

func switchAccountHandler(client google.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		email, err := client.SwitchAccount(ctx)
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr("account="+email, nil)
	}
}
