package actor

import (
	"context"
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
	case runtime.KindPlaylistAdd:
		return playlist(cmd, client, found)
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

func playlist(cmd runtime.Command, client spotify.Client, found *catalog) runtime.Result {
	action := strings.TrimSpace(cmd.Args.Action)
	if action == "" {
		action = "add"
	}
	switch action {
	case "list":
		list, err := client.ListPlaylists(cmd.Ctx)
		if err != nil {
			return runtime.Result{Err: err}
		}
		return runtime.Result{Text: formatPlaylists(list)}
	case "create":
		pl, err := client.CreatePlaylist(cmd.Ctx, cmd.Args.PlaylistName, cmd.Args.Description)
		if err != nil {
			return runtime.Result{Err: err}
		}
		return runtime.Result{Text: fmt.Sprintf("id=%s name=%s uri=%s", pl.ID, pl.Name, pl.URI)}
	case "add":
		pl, err := findPlaylist(cmd.Ctx, client, cmd.Args.PlaylistName)
		if err != nil {
			return runtime.Result{Err: err}
		}
		uri, err := trackURI(cmd, client, found)
		if err != nil {
			return runtime.Result{Err: err}
		}
		if err := client.AddToPlaylist(cmd.Ctx, pl.ID, uri); err != nil {
			return runtime.Result{Err: err}
		}
		return runtime.Result{Text: fmt.Sprintf("added uri=%s to playlist=%s id=%s", uri, pl.Name, pl.ID)}
	case "play":
		pl, err := findPlaylist(cmd.Ctx, client, cmd.Args.PlaylistName)
		if err != nil {
			return runtime.Result{Err: err}
		}
		uri := pl.URI
		if uri == "" {
			uri = "spotify:playlist:" + pl.ID
		}
		if err := client.PlayPlaylist(cmd.Ctx, uri); err != nil {
			return runtime.Result{Err: err}
		}
		bar, err := client.NowPlaying(cmd.Ctx)
		if err != nil {
			return runtime.Result{Err: err}
		}
		return runtime.Result{Text: formatBar(bar)}
	case "update":
		pl, err := findPlaylist(cmd.Ctx, client, cmd.Args.PlaylistName)
		if err != nil {
			return runtime.Result{Err: err}
		}
		if err := client.UpdatePlaylistMetadata(cmd.Ctx, pl.ID, pl.Name, cmd.Args.Description); err != nil {
			return runtime.Result{Err: err}
		}
		return runtime.Result{Text: fmt.Sprintf("updated id=%s name=%s description=%s", pl.ID, pl.Name, cmd.Args.Description)}
	case "delete":
		pl, err := findPlaylist(cmd.Ctx, client, cmd.Args.PlaylistName)
		if err != nil {
			return runtime.Result{Err: err}
		}
		if err := client.DeletePlaylist(cmd.Ctx, pl.ID); err != nil {
			return runtime.Result{Err: err}
		}
		return runtime.Result{Text: fmt.Sprintf("deleted(unfollowed) id=%s name=%s", pl.ID, pl.Name)}
	case "remove":
		pl, err := findPlaylist(cmd.Ctx, client, cmd.Args.PlaylistName)
		if err != nil {
			return runtime.Result{Err: err}
		}
		uri, err := trackURI(cmd, client, found)
		if err != nil {
			return runtime.Result{Err: err}
		}
		if err := client.RemoveFromPlaylist(cmd.Ctx, pl.ID, uri); err != nil {
			return runtime.Result{Err: err}
		}
		return runtime.Result{Text: fmt.Sprintf("removed uri=%s from playlist=%s id=%s", uri, pl.Name, pl.ID)}
	default:
		return runtime.Result{Err: fmt.Errorf("action must be list, create, add, play, or update")}
	}
}

func findPlaylist(ctx context.Context, client spotify.Client, name string) (spotify.Playlist, error) {
	list, err := client.ListPlaylists(ctx)
	if err != nil {
		return spotify.Playlist{}, err
	}
	want := strings.ToLower(strings.TrimSpace(name))
	var names []string
	for _, pl := range list {
		names = append(names, pl.Name)
		if strings.ToLower(pl.Name) == want {
			return pl, nil
		}
	}
	return spotify.Playlist{}, fmt.Errorf("playlist not found; visible=%s", strings.Join(names, ", "))
}

func trackURI(cmd runtime.Command, client spotify.Client, found *catalog) (string, error) {
	if cmd.Args.UseCurrent {
		bar, err := client.NowPlaying(cmd.Ctx)
		if err != nil {
			return "", err
		}
		if bar.URI == "" {
			return "", fmt.Errorf("nothing_playing")
		}
		return bar.URI, nil
	}
	hit, err := pickHit(cmd, client, found)
	if err != nil {
		return "", err
	}
	return hit.URI, nil
}

func formatPlaylists(list []spotify.Playlist) string {
	lines := make([]string, 0, len(list))
	for _, pl := range list {
		lines = append(lines, fmt.Sprintf("id=%s name=%s uri=%s", pl.ID, pl.Name, pl.URI))
	}
	if len(lines) == 0 {
		return "no playlists"
	}
	return strings.Join(lines, "\n")
}

func publish(bus *runtime.Bus, event runtime.SessionEvent) {
	select {
	case bus.SessionEventsChannel <- event:
	case <-time.After(2 * time.Second):
		log.Printf("session event not delivered: %s", event.Type)
	}
}
