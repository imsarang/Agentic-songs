package cache

import (
	"fmt"
	"sync"

	"shanks/internal/google"
)

// Drive holds rows from the latest Drive search (in-process only).
type Drive struct {
	mu   sync.Mutex
	id   int
	rows []google.DriveFile
}

func (s *Drive) Store(rows []google.DriveFile) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id++
	s.rows = append([]google.DriveFile(nil), rows...)
	return s.id
}

func (s *Drive) Pick(searchID, index int) (google.DriveFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.rows) == 0 || (searchID > 0 && searchID != s.id) {
		return google.DriveFile{}, fmt.Errorf("stale handle, search again")
	}
	for _, row := range s.rows {
		if row.Index == index {
			return row, nil
		}
	}
	return google.DriveFile{}, fmt.Errorf("stale handle, search again")
}

func (s *Drive) PickMany(searchID int, indexes []int) ([]google.DriveFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.rows) == 0 || (searchID > 0 && searchID != s.id) {
		return nil, fmt.Errorf("stale handle, search again")
	}
	if len(indexes) == 0 {
		return append([]google.DriveFile(nil), s.rows...), nil
	}
	out := make([]google.DriveFile, 0, len(indexes))
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
