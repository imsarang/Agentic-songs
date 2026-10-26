package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
	"shanks/internal/google/cache"
)

func deleteDraftsHandler(client google.Client, found *cache.Search) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		searchID := req.GetInt("search_id", 0)
		if searchID <= 0 {
			return textOrErr("", fmt.Errorf("search_id is required"))
		}
		indexes, err := resultIndexes(req.GetString("result_indexes", ""))
		if err != nil {
			return textOrErr("", err)
		}
		rows, err := found.PickMany(searchID, indexes)
		if err != nil {
			return textOrErr("", err)
		}
		ids := make([]string, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		if err := client.DeleteDrafts(ctx, ids); err != nil {
			return textOrErr("", err)
		}
		return textOrErr(fmt.Sprintf("deleted %d drafts", len(ids)), nil)
	}
}

func resultIndexes(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var out []int
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' }) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("bad result index %q", part)
		}
		out = append(out, n)
	}
	return out, nil
}
