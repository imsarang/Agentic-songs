package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"shanks/internal/spotify"
)

// Search holds handles from the latest spotify_search (in-process only).
type Search struct {
	mu   sync.Mutex
	id   int
	rows []spotify.Hit
}

func (s *Search) Store(query string, hits []spotify.Hit) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id++
	s.rows = hits
	_ = query
	return s.id
}

func (s *Search) Pick(ctx context.Context, client spotify.Client, query string, searchID, index int, title string) (spotify.Hit, error) {
	if strings.TrimSpace(query) != "" {
		hits, err := client.Search(ctx, query, 1)
		if err != nil {
			return spotify.Hit{}, err
		}
		if len(hits) == 0 {
			return spotify.Hit{}, fmt.Errorf("no tracks")
		}
		s.Store(query, hits)
		return hits[0], nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.rows) == 0 || searchID != s.id {
		return spotify.Hit{}, fmt.Errorf("stale handle, search again")
	}
	for _, row := range s.rows {
		if row.Index != index {
			continue
		}
		if !sameText(row.Title, title) {
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
