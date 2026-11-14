package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/ecommerce/config"
	"shanks/internal/ecommerce/providers"
	"shanks/internal/ecommerce/tools"
	"shanks/internal/transport"
)

func main() {
	log.SetOutput(os.Stderr)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if _, err := config.Load(); err != nil {
		return err
	}

	client := providers.NewHTTPClient()
	boot, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	client.RelevantDomains(boot, config.Electronics())
	cancel()

	log.Printf("Ecommerce MCP up")
	transport.Serve("ecommerce", "0.1.0", func(s *server.MCPServer) {
		tools.Register(s, client)
	})
	return nil
}
