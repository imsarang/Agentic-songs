package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
)

func searchHandler(bus *runtime.Bus) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := waitReady(ctx, bus); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		query := strings.TrimSpace(req.GetString("query", ""))

		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}
		return reply(call(ctx, bus, runtime.KindSearch, runtime.Args{
			Query: query,
			Limit: req.GetInt("limit", 5),
		}))
	}
}
