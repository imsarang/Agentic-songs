package browser

import (
	"fmt"
	"log"
	"time"

	"shanks/internal/config"
	"shanks/internal/runtime"
	"shanks/internal/spotify"
)

func Run(cfg config.Config, bus *runtime.Bus) {
	defer close(bus.ActorStoppedChannel)

	var current *launched
	defer func() { current.close() }()

	for cmd := range bus.CommandsChannel {
		if current == nil || closedPage(current.page) {
			next, err := launch(&cfg)
			if err != nil {
				cmd.Send(runtime.Result{Err: fmt.Errorf("chromium launch: %w", err)})
				continue
			}
			current = next
		}

		result, crashed := dispatch(cmd, current, cfg, bus)
		if crashed {
			current.close()
			current = nil
			publish(bus, runtime.SessionEvent{Type: runtime.EventBrowserCrashed, Err: result.Err})
		}
		cmd.Send(result)
	}
}

func dispatch(cmd runtime.Command, current *launched, cfg config.Config, bus *runtime.Bus) (runtime.Result, bool) {
	switch cmd.Kind {
	case runtime.KindEnsureSession:
		return ensureSession(cmd, current, cfg, bus)
	default:
		return runtime.Result{Err: fmt.Errorf("no browser command for %s yet", cmd.Kind)}, false
	}
}

func ensureSession(cmd runtime.Command, current *launched, cfg config.Config, bus *runtime.Bus) (runtime.Result, bool) {
	view, err := spotify.OpenPlayer(current.page)
	if browserClosed(err) || closedPage(current.page) {
		return runtime.Result{Err: err}, true
	}
	if err != nil {
		return runtime.Result{Err: err}, false
	}
	log.Printf("spotify view=%s url=%s", view, current.page.URL())

	if view != spotify.ViewPlayer {
		err = waitForManualLogin(cmd.Ctx, current.page, cfg.ProfileDir, func(event runtime.SessionEvent) {
			publish(bus, event)
		})
		if err != nil {
			return runtime.Result{Err: err}, browserClosed(err) || closedPage(current.page)
		}
	}
	if err = saveSession(current.page, cfg.ProfileDir); err != nil {
		return runtime.Result{Err: err}, false
	}
	publish(bus, runtime.SessionEvent{Type: runtime.EventLoggedIn, Phase: runtime.PhaseReady})
	return runtime.Result{Phase: runtime.PhaseReady, Text: "player shell visible"}, false
}

func publish(bus *runtime.Bus, event runtime.SessionEvent) {
	select {
	case bus.SessionEventsChannel <- event:
	case <-time.After(2 * time.Second):
		log.Printf("session event not delivered: %s", event.Type)
	}
}
