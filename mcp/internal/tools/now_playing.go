package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
)

func nowPlayingHandler(bus *runtime.Bus) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := waitReady(ctx, bus); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return reply(call(ctx, bus, runtime.KindNowPlaying, runtime.Args{}))
	}
}
