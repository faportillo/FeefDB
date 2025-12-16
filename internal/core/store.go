package core

import (
	"fmt"
	"sync"
)

type Store struct {
	mu sync.RWMutex
	cols map[string]*Collection
}

func NewStore() *Store {
	return &Store{
		cols: make(map[string]*Collection),
	}
}

func (s *Store) CreateCollection(name string, dim int, metric DistanceMetric) (*Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cols[name]; exists {
		return nil, fmt.Errorf("collection already exists: %q", name)
	}

	c, err := NewCollection(name, dim, metric)
	if err != nil {
		return nil, err
	}

	s.cols[name] = c
	return c, nil
}

func (s *Store) GetCollection(name string) (*Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c := s.cols[name]
	if c == nil {
		return nil, fmt.Errorf("collection not found: %q", name)
	}
	return c, nil
}

// Optional helper if you want later:
func (s *Store) ListCollections() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, 0, len(s.cols))
	for name := range s.cols {
		out = append(out, name)
	}
	return out
}