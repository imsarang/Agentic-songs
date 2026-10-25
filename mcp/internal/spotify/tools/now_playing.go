package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/spotify"
)

func nowPlayingHandler(client spotify.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		bar, err := client.NowPlaying(ctx)
		if err != nil {
			return textOrErr("", err)
		}
		if bar.Title == "" {
			return textOrErr("", fmt.Errorf("nothing_playing"))
		}
		return textOrErr(spotify.FormatBar(bar), nil)
	}
}
