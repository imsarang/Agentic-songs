package main

import (
	"context"
	"log"
	"os"
	"shanks/internal/browser"
	"shanks/internal/config"
	"shanks/internal/runtime"
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
	log.Printf("account=%s profile=%s headless=%t", config.RedactedEmail(cfg.GoogleEmail), cfg.ProfileDir, cfg.Headless)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := runtime.BusInit()
	supervisorStopped := make(chan struct{})

	go runtime.Signals(cancel)
	go browser.Run(cfg, bus)
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
