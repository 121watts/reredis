package store

import (
	"sync"
)

type Store struct {
	data map[string]string
	mu   sync.Mutex
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]

	return v, ok
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if ok {
		delete(s.data, key)
	}

	return ok
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}
