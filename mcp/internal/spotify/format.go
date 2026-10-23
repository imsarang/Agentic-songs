package spotify

import (
	"context"
	"fmt"
	"strings"
)

func FormatHits(id int, hits []Hit) string {
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

func FormatBar(bar Bar) string {
	return fmt.Sprintf("title=%s artist=%s uri=%s", bar.Title, bar.Artist, bar.URI)
}

func FormatPlaylists(list []Playlist) string {
	lines := make([]string, 0, len(list))
	for _, pl := range list {
		lines = append(lines, fmt.Sprintf("id=%s name=%s uri=%s", pl.ID, pl.Name, pl.URI))
	}
	if len(lines) == 0 {
		return "no playlists"
	}
	return strings.Join(lines, "\n")
}

func WithAnalytics(ctx context.Context, client Client, bar Bar) string {
	text := FormatBar(bar)
	id := trackIDFromURI(bar.URI)
	if id == "" {
		return text + "\nanalytics: unavailable (no track id on now-playing)"
	}
	feat, err := client.AudioFeatures(ctx, id)
	if err != nil {
		return text + "\nanalytics: unavailable — Spotify blocks /audio-features for this app (403). That endpoint is deprecated for Development Mode."
	}
	return text + "\n" + provideAnalytics(feat)
}

func trackIDFromURI(uri string) string {
	const p = "spotify:track:"
	if !strings.HasPrefix(uri, p) {
		return ""
	}
	return strings.TrimPrefix(uri, p)
}

func provideAnalytics(f TrackFeatures) string {
	mode := "minor"
	if f.Mode == 1 {
		mode = "major"
	}
	return strings.Join([]string{
		"analytics (present to the user as a short summary or bullet list; do not dump raw keys):",
		fmt.Sprintf("- acousticness: %.3f (0=electric … 1=acoustic)", f.Acousticness),
		fmt.Sprintf("- danceability: %.3f", f.Danceability),
		fmt.Sprintf("- energy: %.3f", f.Energy),
		fmt.Sprintf("- valence: %.3f (0=sad … 1=happy)", f.Valence),
		fmt.Sprintf("- tempo: %.1f bpm", f.Tempo),
		fmt.Sprintf("- key: %d  mode: %s", f.Key, mode),
		fmt.Sprintf("- loudness: %.1f dB", f.Loudness),
		fmt.Sprintf("- speechiness: %.3f", f.Speechiness),
		fmt.Sprintf("- liveness: %.3f", f.Liveness),
		fmt.Sprintf("- time_signature: %.0f", f.TimeSignature),
	}, "\n")
}

func FindPlaylist(ctx context.Context, client Client, name string) (Playlist, error) {
	list, err := client.ListPlaylists(ctx)
	if err != nil {
		return Playlist{}, err
	}
	want := strings.ToLower(strings.TrimSpace(name))
	var names []string
	for _, pl := range list {
		names = append(names, pl.Name)
		if strings.ToLower(pl.Name) == want {
			return pl, nil
		}
	}
	return Playlist{}, fmt.Errorf("playlist not found; visible=%s", strings.Join(names, ", "))
}
