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
