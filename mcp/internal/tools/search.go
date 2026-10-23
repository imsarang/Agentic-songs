package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/cache"
	"shanks/internal/spotify"
)

func searchHandler(client spotify.Client, found *cache.Search) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		query := strings.TrimSpace(req.GetString("query", ""))
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}
		hits, err := client.Search(ctx, query, req.GetInt("limit", 5))
		if err != nil {
			return textOrErr("", err)
		}
		id := found.Store(query, hits)
		return textOrErr(spotify.FormatHits(id, hits), nil)
	}
}
