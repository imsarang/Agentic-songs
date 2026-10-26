package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
	"shanks/internal/google/cache"
)

var errNeedMessage = fmt.Errorf("message_id or search_id is required")

func deleteMailHandler(client google.Client, found *cache.Search) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		ids, err := mailIDs(found, req)
		if err != nil {
			return textOrErr("", err)
		}
		for _, id := range ids {
			if err := client.DeleteMails(ctx, id); err != nil {
				return textOrErr("", err)
			}
		}
		return textOrErr(fmt.Sprintf("trashed %d", len(ids)), nil)
	}
}

func mailIDs(found *cache.Search, req mcp.CallToolRequest) ([]string, error) {
	if id := strings.TrimSpace(req.GetString("message_id", "")); id != "" {
		return []string{id}, nil
	}
	searchID := req.GetInt("search_id", 0)
	if searchID <= 0 {
		return nil, errNeedMessage
	}
	var indexes []int
	if raw := strings.TrimSpace(req.GetString("result_indexes", "")); raw != "" {
		parsed, err := resultIndexes(raw)
		if err != nil {
			return nil, err
		}
		indexes = parsed
	} else if idx := req.GetInt("result_index", -1); idx >= 0 {
		indexes = []int{idx}
	}
	rows, err := found.PickMany(searchID, indexes)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}
