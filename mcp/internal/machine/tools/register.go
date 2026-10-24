package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Register(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("machine_spec",
		mcp.WithDescription("Read this machine's identity: vendor, product, CPU, memory, filesystems, uptime, and whether it is physical or virtual. Returns one JSON object."),
	), specHandler())
}
