package transport

import (
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

const (
	envTransport = "MCP_TRANSPORT" // stdio (default) | http
	envAddr      = "MCP_ADDR"      // listen addr for http, default :8080
)

// Serve builds an MCP server, lets the use case register its tools, and
// listens on stdio or streamable HTTP. The transport does not know the tools.
func Serve(name, version string, register func(*server.MCPServer)) {
	s := server.NewMCPServer(name, version, server.WithToolCapabilities(true))
	register(s)

	switch strings.ToLower(strings.TrimSpace(os.Getenv(envTransport))) {
	case "http":
		serveHTTP(s)
	default:
		serveStdio(s)
	}
}

func serveStdio(s *server.MCPServer) {
	if err := server.ServeStdio(s); err != nil {
		log.Printf("mcp stopped: %v", err)
		return
	}
	log.Printf("mcp stopped: stdin closed")
}

func serveHTTP(s *server.MCPServer) {
	addr := strings.TrimSpace(os.Getenv(envAddr))
	if addr == "" {
		addr = ":8080"
	}
	httpServer := server.NewStreamableHTTPServer(
		s,
		server.WithEndpointPath("/mcp"),
		server.WithStateLess(true),
		server.WithDisableLocalhostProtection(true),
	)
	log.Printf("mcp http transport addr=%s path=/mcp", addr)
	if err := httpServer.Start(addr); err != nil {
		log.Printf("mcp stopped: %v", err)
	}
}
