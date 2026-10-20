package mcp

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
	"shanks/internal/tools"
)

func Serve(bus *runtime.Bus) {
	s := server.NewMCPServer("spotify", "0.1.0", server.WithToolCapabilities(true))
	tools.Register(s, bus)
	err := server.ServeStdio(s)
	if err != nil {
		log.Printf("mcp stopped: %v", err)
		return
	}
	log.Printf("mcp stopped: stdin closed")
}
