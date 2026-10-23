package main

import (
	"log"
	"os"

	"shanks/internal/auth"
	"shanks/internal/config"
	"shanks/internal/mcp"
	"shanks/internal/spotify"
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
	mcp.Serve(client)
	return nil
}
