package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
)

func createDraftHandler(client google.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		draft, err := client.CreateDraft(ctx,
			req.GetString("to", ""),
			req.GetString("subject", ""),
			req.GetString("body", ""),
		)
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr(google.FormatDraft(draft), nil)
	}
}
