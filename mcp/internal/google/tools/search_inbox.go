package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
	"shanks/internal/google/cache"
)

func searchBox(client google.Client, found *cache.Search, kind google.EmailType) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		rows, err := client.SearchMails(ctx, kind, req.GetString("query", ""), req.GetInt("limit", 10))
		if err != nil {
			return textOrErr("", err)
		}
		id := found.Store(rows)
		return textOrErr(google.FormatMails(id, rows), nil)
	}
}
