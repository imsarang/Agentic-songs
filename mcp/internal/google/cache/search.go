package cache

import (
	"fmt"
	"sync"

	"shanks/internal/google"
)

// Search holds rows from the latest mailbox search (in-process only).
type Search struct {
	mu   sync.Mutex
	id   int
	rows []google.Mail
}

func (s *Search) Store(rows []google.Mail) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id++
	s.rows = append([]google.Mail(nil), rows...)
	return s.id
}

func (s *Search) Pick(searchID, index int) (google.Mail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.rows) == 0 || (searchID > 0 && searchID != s.id) {
		return google.Mail{}, fmt.Errorf("stale handle, search again")
	}
	for _, row := range s.rows {
		if row.Index == index {
			return row, nil
		}
	}
	return google.Mail{}, fmt.Errorf("stale handle, search again")
}

func (s *Search) PickMany(searchID int, indexes []int) ([]google.Mail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.rows) == 0 || (searchID > 0 && searchID != s.id) {
		return nil, fmt.Errorf("stale handle, search again")
	}
	if len(indexes) == 0 {
		return append([]google.Mail(nil), s.rows...), nil
	}
	out := make([]google.Mail, 0, len(indexes))
	for _, index := range indexes {
		found := false
		for _, row := range s.rows {
			if row.Index == index {
				out = append(out, row)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("stale handle, search again")
		}
	}
	return out, nil
}
