package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/spotify"
)

func controlHandler(client spotify.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		action := strings.TrimSpace(req.GetString("action", ""))
		switch action {
		case "pause", "resume", "next", "previous":
		default:
			return mcp.NewToolResultError("action must be pause, resume, next, or previous"), nil
		}
		if err := client.Control(ctx, action); err != nil {
			return textOrErr("", err)
		}
		bar, err := client.NowPlaying(ctx)
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr(spotify.WithAnalytics(ctx, client, bar), nil)
	}
}
