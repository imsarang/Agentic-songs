package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
)

func sendMailHandler(client google.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		mail, err := client.SendMail(ctx, req.GetString("draft_id", ""))
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr("sent id="+mail.ID, nil)
	}
}
