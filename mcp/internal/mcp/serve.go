package mcp

import (
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
	"shanks/internal/tools"
)

const (
	envTransport = "MCP_TRANSPORT" // stdio (default) | http
	envAddr      = "MCP_ADDR"      // listen addr for http, default :8080
)

func Serve(bus *runtime.Bus) {
	s := server.NewMCPServer("spotify", "0.1.0", server.WithToolCapabilities(true))
	tools.Register(s, bus)

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
		// allow calls via docker service hostname, not only localhost
		server.WithDisableLocalhostProtection(true),
	)
	log.Printf("mcp http transport addr=%s path=/mcp", addr)
	if err := httpServer.Start(addr); err != nil {
		log.Printf("mcp stopped: %v", err)
	}
}
