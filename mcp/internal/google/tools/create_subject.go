package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
)

func createSubjectHandler(client google.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		subject := strings.TrimSpace(req.GetString("subject", ""))
		if err := client.CreateSubject(ctx, subject); err != nil {
			return textOrErr("", err)
		}
		return textOrErr("subject="+subject, nil)
	}
}
