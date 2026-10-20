package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
)

func controlHandler(bus *runtime.Bus) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := waitReady(ctx, bus); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		action := strings.TrimSpace(req.GetString("action", ""))
		switch action {
		case "pause", "resume", "next", "previous":
		default:
			return mcp.NewToolResultError("action must be pause, resume, next, or previous"), nil
		}
		return reply(call(ctx, bus, runtime.KindControl, runtime.Args{Action: action}))
	}
}
