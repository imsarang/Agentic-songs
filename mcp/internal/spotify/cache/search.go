package cache

import (
	"context"
	"fmt"
	"log"
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
	q := strings.TrimSpace(query)
	title = strings.TrimSpace(title)

	// Prefer handles from the latest search when the agent passes them.
	if title != "" || searchID > 0 {
		s.mu.Lock()
		rows := append([]spotify.Hit(nil), s.rows...)
		id := s.id
		s.mu.Unlock()
		if len(rows) == 0 {
			return spotify.Hit{}, fmt.Errorf("stale handle, search again")
		}
		if searchID > 0 && searchID != id {
			return spotify.Hit{}, fmt.Errorf("stale handle, search again")
		}
		for _, row := range rows {
			if title != "" {
				if sameText(row.Title, title) || strings.Contains(normalize(row.Title), normalize(title)) {
					if index > 0 && row.Index != index {
						continue
					}
					log.Printf("play pick handle index=%d title=%q uri=%s", row.Index, row.Title, row.URI)
					return row, nil
				}
				continue
			}
			if row.Index == index {
				log.Printf("play pick handle index=%d title=%q uri=%s", row.Index, row.Title, row.URI)
				return row, nil
			}
		}
		if q == "" {
			return spotify.Hit{}, fmt.Errorf("stale handle, search again")
		}
	}

	if q == "" {
		return spotify.Hit{}, fmt.Errorf("pass a query or a search handle")
	}

	// Never use limit=1 — Spotify's #1 is often unrelated. Score titles.
	hits, err := client.Search(ctx, q, 8)
	if err != nil {
		return spotify.Hit{}, err
	}
	if len(hits) == 0 {
		return spotify.Hit{}, fmt.Errorf("no tracks")
	}
	s.Store(q, hits)
	hit := bestHit(q, hits)
	log.Printf("play pick query=%q chose title=%q artist=%q uri=%s", q, hit.Title, hit.Artist, hit.URI)
	return hit, nil
}

func bestHit(query string, hits []spotify.Hit) spotify.Hit {
	q := normalize(query)
	best := hits[0]
	bestScore := titleScore(q, normalize(best.Title))
	for _, h := range hits[1:] {
		score := titleScore(q, normalize(h.Title))
		if score > bestScore {
			bestScore = score
			best = h
		}
	}
	return best
}

func titleScore(query, title string) int {
	if query == "" || title == "" {
		return 0
	}
	if title == query {
		return 100
	}
	base := title
	if i := strings.Index(title, " from "); i > 0 {
		base = strings.TrimSpace(title[:i])
	}
	if base == query {
		return 95
	}
	if strings.HasPrefix(title, query) || strings.HasPrefix(query, base) {
		return 85
	}
	if strings.Contains(title, query) {
		return 70
	}
	qt := strings.Fields(query)
	matched := 0
	for _, t := range qt {
		if len(t) >= 2 && strings.Contains(title, t) {
			matched++
		}
	}
	if len(qt) > 0 && matched == len(qt) {
		return 50 + matched
	}
	if len(qt) > 0 && matched*2 >= len(qt) {
		return 20 + matched
	}
	return 0
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, r := range []string{"(", ")", "\"", "'", "[", "]", "—", "-"} {
		s = strings.ReplaceAll(s, r, " ")
	}
	return strings.Join(strings.Fields(s), " ")
}

func sameText(a, b string) bool {
	return normalize(a) != "" && normalize(a) == normalize(b)
}
