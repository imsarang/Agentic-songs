package main

import (
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/spotify"
	"shanks/internal/spotify/auth"
	"shanks/internal/spotify/config"
	"shanks/internal/spotify/tools"
	"shanks/internal/transport"
)

func main() {
	log.SetOutput(os.Stderr)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log.Printf("spotify api client_id=%s data=%s", config.RedactedClientID(cfg.ClientID), cfg.DataDir)

	tokens := auth.NewSource(auth.Path(cfg.DataDir), cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)
	client := spotify.NewHTTPClient(tokens)
	transport.Serve("spotify", "0.1.0", func(s *server.MCPServer) {
		tools.Register(s, client)
	})
	return nil
}
