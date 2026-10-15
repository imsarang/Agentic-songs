package runtime

import (
	"context"
	"log"
)

func Orchestrate(ctx context.Context, bus *Bus) {
	phase := PhaseStarting
	pauseLogin := false
	loginInFlight := false
	var reply <-chan Result

	enqueue := func() {
		if pauseLogin || loginInFlight {
			return
		}

		command := CommandInit(ctx, KindEnsureSession, Args{})
		select {
		case bus.CommandsChannel <- command:
			reply = command.ReplyChannel
			loginInFlight = true
		case <-ctx.Done():
		}
	}

	enqueue()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-bus.SessionEventsChannel:
			switch event.Type {

			case EventNeedsLogin:
				phase = PhaseLoggingIn

			case EventLoggedIn:
				phase = PhaseReady
				pauseLogin = false
				bus.MarkReady()

			case EventChallenge:
				phase = PhaseChallenge
				pauseLogin = true

			case EventExpired:
				phase = PhaseExpired
				enqueue()

			case EventBrowserCrashed:
				loginInFlight = false
				reply = nil
				enqueue()
			}
			log.Printf("session event=%s phase=%s", event.Type, phase)

		case result := <-reply:
			loginInFlight = false
			reply = nil
			if result.Err != nil {
				log.Printf("ensureSession error: %v", result.Err)
				continue
			}
			if result.Phase != "" {
				phase = result.Phase
			}
			log.Printf("ensureSession phase=%s", phase)
		}
	}
}
