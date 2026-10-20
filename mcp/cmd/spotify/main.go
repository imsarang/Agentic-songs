package main

import (
	"context"
	"log"
	"os"

	"shanks/internal/actor"
	"shanks/internal/auth"
	"shanks/internal/config"
	"shanks/internal/mcp"
	"shanks/internal/runtime"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := runtime.BusInit()
	supervisorStopped := make(chan struct{})

	go runtime.Signals(cancel)
	go actor.Run(cfg, bus, client)
	go func() {
		mcp.Serve(bus)
		cancel()
	}()
	go func() {
		defer close(supervisorStopped)
		runtime.Orchestrate(ctx, bus)
	}()

	<-ctx.Done()
	<-supervisorStopped
	close(bus.CommandsChannel)
	<-bus.ActorStoppedChannel
	log.Printf("actor stopped")
	return nil
}
