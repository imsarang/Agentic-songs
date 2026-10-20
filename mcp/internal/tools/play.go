package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
)

func playHandler(bus *runtime.Bus) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := waitReady(ctx, bus); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		query := strings.TrimSpace(req.GetString("query", ""))
		title := strings.TrimSpace(req.GetString("title", ""))
		if query == "" && title == "" {
			return mcp.NewToolResultError("pass a query or a search handle"), nil
		}

		args := runtime.Args{Query: query}
		if query == "" {
			args = runtime.Args{
				Title:       title,
				Artist:      strings.TrimSpace(req.GetString("artist", "")),
				ResultIndex: req.GetInt("result_index", 0),
				SearchID:    req.GetInt("search_id", 0),
			}
		}
		return reply(call(ctx, bus, runtime.KindPlay, args))
	}
}
