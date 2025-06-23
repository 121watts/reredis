package store

import (
	"container/list"
	"sync"
)

type cacheItem struct {
	key   string
	value string
}

type Store struct {
	data    map[string]*list.Element // map key to list element
	lruList *list.List               // doubly linked list for LRU ordering
	maxSize int                      // maximum number of items
	mu      sync.Mutex
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key already exists
	if elem, exists := s.data[key]; exists {
		// Update existing item and move to front
		elem.Value.(*cacheItem).value = value
		s.lruList.MoveToFront(elem)
		return
	}

	// Add new item
	item := &cacheItem{key: key, value: value}
	elem := s.lruList.PushFront(item)
	s.data[key] = elem

	// Evict if over capacity
	if s.lruList.Len() > s.maxSize {
		s.evictLRU()
	}
}

func (s *Store) evictLRU() {
	// Remove least recently used item (back of list)
	elem := s.lruList.Back()
	if elem != nil {
		item := elem.Value.(*cacheItem)
		delete(s.data, item.key)
		s.lruList.Remove(elem)
	}
}

func (s *Store) GetAll() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	dataCopy := make(map[string]string, len(s.data))
	for key, elem := range s.data {
		item := elem.Value.(*cacheItem)
		dataCopy[key] = item.value
	}

	return dataCopy
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	elem, ok := s.data[key]
	if !ok {
		return "", false
	}

	// Move to front (most recently used)
	s.lruList.MoveToFront(elem)
	item := elem.Value.(*cacheItem)
	return item.value, true
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	elem, ok := s.data[key]
	if ok {
		delete(s.data, key)
		s.lruList.Remove(elem)
	}

	return ok
}

func NewStore() *Store {
	return &Store{
		data:    make(map[string]*list.Element),
		lruList: list.New(),
		maxSize: 1000, // Default max size, could be configurable
	}
}
