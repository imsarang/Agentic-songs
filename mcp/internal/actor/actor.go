package actor

import (
	"fmt"
	"log"
	"strings"
	"time"

	"shanks/internal/config"
	"shanks/internal/runtime"
	"shanks/internal/spotify"
)

type catalog struct {
	id    int
	query string
	rows  []spotify.Hit
}

// Run reads commands and calls the Spotify API client. No browser.
func Run(cfg config.Config, bus *runtime.Bus, client spotify.Client) {
	defer close(bus.ActorStoppedChannel)

	var found catalog
	phase := runtime.PhaseStarting

	for cmd := range bus.CommandsChannel {
		log.Printf("command kind=%s", cmd.Kind)
		started := time.Now()
		result := dispatch(cmd, cfg, bus, client, &found, &phase)
		if result.Phase != "" {
			phase = result.Phase
		}
		cmd.Send(result)
		log.Printf("command done kind=%s ms=%d err=%v", cmd.Kind, time.Since(started).Milliseconds(), result.Err)
	}
}

func dispatch(cmd runtime.Command, cfg config.Config, bus *runtime.Bus, client spotify.Client, found *catalog, phase *runtime.Phase) runtime.Result {
	switch cmd.Kind {
	case runtime.KindSession:
		return runtime.Result{
			Phase: *phase,
			Text:  config.RedactedClientID(cfg.ClientID),
		}
	case runtime.KindEnsureSession:
		return ensureAuth(cmd, bus, client, phase)
	case runtime.KindSearch:
		return search(cmd, client, found)
	case runtime.KindPlay:
		return play(cmd, client, found)
	case runtime.KindControl:
		return control(cmd, client)
	case runtime.KindNowPlaying:
		return nowPlaying(cmd, client)
	default:
		return runtime.Result{Err: fmt.Errorf("no command for %s", cmd.Kind)}
	}
}

func ensureAuth(cmd runtime.Command, bus *runtime.Bus, client spotify.Client, phase *runtime.Phase) runtime.Result {
	if err := client.EnsureAuth(cmd.Ctx); err != nil {
		publish(bus, runtime.SessionEvent{Type: runtime.EventNeedsLogin, Err: err})
		return runtime.Result{Err: err, Phase: runtime.PhaseLoggingIn}
	}
	*phase = runtime.PhaseReady
	publish(bus, runtime.SessionEvent{Type: runtime.EventLoggedIn, Phase: runtime.PhaseReady})
	return runtime.Result{Phase: runtime.PhaseReady, Text: "spotify api token ready"}
}

func search(cmd runtime.Command, client spotify.Client, found *catalog) runtime.Result {
	hits, err := client.Search(cmd.Ctx, cmd.Args.Query, cmd.Args.Limit)
	if err != nil {
		return runtime.Result{Err: err}
	}
	found.id++
	found.query = cmd.Args.Query
	found.rows = hits
	return runtime.Result{Text: formatHits(found.id, hits)}
}

func play(cmd runtime.Command, client spotify.Client, found *catalog) runtime.Result {
	hit, err := pickHit(cmd, client, found)
	if err != nil {
		return runtime.Result{Err: err}
	}
	if err := client.Play(cmd.Ctx, hit.URI); err != nil {
		return runtime.Result{Err: err}
	}
	bar, err := client.NowPlaying(cmd.Ctx)
	if err != nil {
		return runtime.Result{Err: err}
	}
	return runtime.Result{Text: formatBar(bar)}
}

func control(cmd runtime.Command, client spotify.Client) runtime.Result {
	if err := client.Control(cmd.Ctx, cmd.Args.Action); err != nil {
		return runtime.Result{Err: err}
	}
	bar, err := client.NowPlaying(cmd.Ctx)
	if err != nil {
		return runtime.Result{Err: err}
	}
	return runtime.Result{Text: formatBar(bar)}
}

func nowPlaying(cmd runtime.Command, client spotify.Client) runtime.Result {
	bar, err := client.NowPlaying(cmd.Ctx)
	if err != nil {
		return runtime.Result{Err: err}
	}
	if bar.Title == "" {
		return runtime.Result{Err: fmt.Errorf("nothing_playing")}
	}
	return runtime.Result{Text: formatBar(bar)}
}

func pickHit(cmd runtime.Command, client spotify.Client, found *catalog) (spotify.Hit, error) {
	if strings.TrimSpace(cmd.Args.Query) != "" {
		hits, err := client.Search(cmd.Ctx, cmd.Args.Query, 1)
		if err != nil {
			return spotify.Hit{}, err
		}
		if len(hits) == 0 {
			return spotify.Hit{}, fmt.Errorf("no tracks")
		}
		found.id++
		found.query = cmd.Args.Query
		found.rows = hits
		return hits[0], nil
	}
	if found == nil || len(found.rows) == 0 || cmd.Args.SearchID != found.id {
		return spotify.Hit{}, fmt.Errorf("stale handle, search again")
	}
	for _, row := range found.rows {
		if row.Index != cmd.Args.ResultIndex {
			continue
		}
		if !sameText(row.Title, cmd.Args.Title) {
			return spotify.Hit{}, fmt.Errorf("stale handle, search again")
		}
		return row, nil
	}
	return spotify.Hit{}, fmt.Errorf("stale handle, search again")
}

func sameText(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	return a != "" && a == b
}

func formatHits(id int, hits []spotify.Hit) string {
	lines := make([]string, 0, len(hits)+1)
	lines = append(lines, fmt.Sprintf("search_id=%d", id))
	for _, hit := range hits {
		lines = append(lines, fmt.Sprintf(
			"index=%d title=%s artist=%s uri=%s",
			hit.Index, hit.Title, hit.Artist, hit.URI,
		))
	}
	return strings.Join(lines, "\n")
}

func formatBar(bar spotify.Bar) string {
	return fmt.Sprintf("title=%s artist=%s uri=%s", bar.Title, bar.Artist, bar.URI)
}

func publish(bus *runtime.Bus, event runtime.SessionEvent) {
	select {
	case bus.SessionEventsChannel <- event:
	case <-time.After(2 * time.Second):
		log.Printf("session event not delivered: %s", event.Type)
	}
}
