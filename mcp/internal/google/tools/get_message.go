package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
	"shanks/internal/google/cache"
)

func getMessageHandler(client google.Client, found *cache.Search) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		id, err := messageID(found, req)
		if err != nil {
			return textOrErr("", err)
		}
		msg, err := client.GetMessage(ctx, id)
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr(google.FormatMessage(msg), nil)
	}
}

func messageID(found *cache.Search, req mcp.CallToolRequest) (string, error) {
	if id := strings.TrimSpace(req.GetString("message_id", "")); id != "" {
		return id, nil
	}
	searchID := req.GetInt("search_id", 0)
	if searchID <= 0 {
		return "", errNeedMessage
	}
	row, err := found.Pick(searchID, req.GetInt("result_index", 0))
	if err != nil {
		return "", err
	}
	return row.ID, nil
}
