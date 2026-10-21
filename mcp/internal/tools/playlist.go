package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
)

func playlistHandler(bus *runtime.Bus) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := waitReady(ctx, bus); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		action := strings.TrimSpace(req.GetString("action", "add"))
		name := strings.TrimSpace(req.GetString("playlist_name", ""))
		if name == "" && action != "list" {
			return mcp.NewToolResultError("playlist_name required"), nil
		}

		useCurrent := req.GetString("use_current", "")
		args := runtime.Args{
			Action:       action,
			PlaylistName: name,
			Description:  strings.TrimSpace(req.GetString("description", "")),
			Query:        strings.TrimSpace(req.GetString("query", "")),
			Title:        strings.TrimSpace(req.GetString("title", "")),
			Artist:       strings.TrimSpace(req.GetString("artist", "")),
			ResultIndex:  req.GetInt("result_index", 0),
			SearchID:     req.GetInt("search_id", 0),
			UseCurrent:   useCurrent == "true" || useCurrent == "1",
		}
		return reply(call(ctx, bus, runtime.KindPlaylistAdd, args))
	}
}
