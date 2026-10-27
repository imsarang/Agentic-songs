package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/spotify"
	"shanks/internal/spotify/cache"
)

func playHandler(client spotify.Client, found *cache.Search) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		query := strings.TrimSpace(req.GetString("query", ""))
		title := strings.TrimSpace(req.GetString("title", ""))
		searchID := req.GetInt("search_id", 0)
		index := req.GetInt("result_index", 0)
		if query == "" && title == "" && searchID <= 0 {
			return mcp.NewToolResultError("pass a query or a search handle"), nil
		}
		hit, err := found.Pick(ctx, client, query, searchID, index, title)
		if err != nil {
			return textOrErr("", err)
		}
		if err := client.Play(ctx, hit.URI); err != nil {
			return textOrErr("", err)
		}
		bar, err := client.NowPlaying(ctx)
		if err != nil {
			return textOrErr(fmt.Sprintf("requested title=%s artist=%s uri=%s\nnow_playing: error: %v", hit.Title, hit.Artist, hit.URI, err), nil)
		}
		out := fmt.Sprintf("requested title=%s artist=%s uri=%s\n%s", hit.Title, hit.Artist, hit.URI, spotify.FormatBar(bar))
		if bar.URI != "" && bar.URI != hit.URI {
			out += "\nmismatch: player bar differs from requested track — desktop Spotify may be hung or on another device; open Spotify and play anything once, then retry"
		}
		return textOrErr(out, nil)
	}
}
