package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/cache"
	"shanks/internal/spotify"
)

func playHandler(client spotify.Client, found *cache.Search) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		query := strings.TrimSpace(req.GetString("query", ""))
		title := strings.TrimSpace(req.GetString("title", ""))
		if query == "" && title == "" {
			return mcp.NewToolResultError("pass a query or a search handle"), nil
		}
		hit, err := found.Pick(ctx, client, query, req.GetInt("search_id", 0), req.GetInt("result_index", 0), title)
		if err != nil {
			return textOrErr("", err)
		}
		if err := client.Play(ctx, hit.URI); err != nil {
			return textOrErr("", err)
		}
		bar, err := client.NowPlaying(ctx)
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr(spotify.WithAnalytics(ctx, client, bar), nil)
	}
}
