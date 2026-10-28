package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
	"shanks/internal/google/cache"
)

func driveHandler(client google.Client, found *cache.Drive) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		switch req.GetString("action", "search") {
		case "search":
			rows, err := client.SearchDrive(ctx, req.GetString("query", ""), 20)
			if err != nil {
				return textOrErr("", err)
			}
			return textOrErr(google.FormatDriveFiles(found.Store(rows), rows), nil)
		case "open":
			f, err := found.Pick(req.GetInt("search_id", 0), req.GetInt("result_index", 0))
			if err != nil {
				return textOrErr("", err)
			}
			body, path, err := client.OpenDrive(ctx, f, req.GetString("download", "") == "true")
			if err != nil {
				return textOrErr("", err)
			}
			shown := body
			if path != "" {
				shown = path + "\n" + shown
			}
			return textOrErr(fmt.Sprintf("opened name=%s mime=%s path=%s\n---\n%s", f.Name, f.Mime, path, shown), nil)
		case "put":
			f, err := client.PutDrive(ctx, req.GetString("kind", "file"), req.GetString("name", ""), req.GetString("parent", ""), req.GetString("path", ""))
			if err != nil {
				return textOrErr("", err)
			}
			return textOrErr(fmt.Sprintf("id=%s name=%s mime=%s", f.ID, f.Name, f.Mime), nil)
		case "delete":
			indexes, err := resultIndexes(req.GetString("result_indexes", ""))
			if err != nil {
				return textOrErr("", err)
			}
			rows, err := found.PickMany(req.GetInt("search_id", 0), indexes)
			if err != nil {
				return textOrErr("", err)
			}
			ids := make([]string, 0, len(rows))
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			if err := client.DeleteDrive(ctx, ids); err != nil {
				return textOrErr("", err)
			}
			return textOrErr(fmt.Sprintf("trashed %d", len(ids)), nil)
		default:
			return textOrErr("", fmt.Errorf("action must be search, open, put, or delete"))
		}
	}
}
