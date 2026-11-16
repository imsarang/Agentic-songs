package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/redis/go-redis/v9"
	"shanks/internal/ecommerce/providers"
)

func Register(s *server.MCPServer, client providers.Client, rdb *redis.Client) {
	s.AddTool(mcp.NewTool("ecommerce_search",
		mcp.WithDescription("Search Indian electronics shops. group is audio, audiophile, keyboards, or diy."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Product text")),
		mcp.WithString("group", mcp.Description("audio, audiophile, keyboards, or diy. Empty searches every live shop.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rows, err := client.SearchProduct(ctx, req.GetString("group", ""), req.GetString("query", ""), rdb)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		var b strings.Builder
		for i, row := range rows {
			fmt.Fprintf(&b, "%d. %s — %s — %s\n", i, row.Title, row.Price, row.URL)
		}
		if b.Len() == 0 {
			return mcp.NewToolResultText("no products"), nil
		}
		return mcp.NewToolResultText(b.String()), nil
	})
}
