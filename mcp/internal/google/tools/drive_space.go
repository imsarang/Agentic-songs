package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
)

func driveSpaceHandler(client google.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		quota, err := client.CheckDriveSpace(ctx)
		if err != nil {
			return textOrErr("", err)
		}
		return textOrErr(google.FormatDrive(quota), nil)
	}
}
