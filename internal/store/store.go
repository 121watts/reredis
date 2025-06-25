package store

import (
	"container/list"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type cacheItem struct {
	key        string
	value      string
	expiration *time.Time
}

type Store struct {
	data    map[string]*list.Element // map key to list element
	lruList *list.List               // doubly linked list for LRU ordering
	maxSize int                      // maximum number of items
	mu      sync.Mutex
	withTTL map[string]bool
}

func (s *Store) Set(key, value string) {
	s.setInternal(key, value, nil)
}

func (s *Store) SetWithTTL(key, value string, ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	s.setInternal(key, value, &expiration)
}

func (s *Store) setInternal(key, value string, expiration *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key already exists
	if elem, exists := s.data[key]; exists {
		// Update existing item and move to front
		elem.Value.(*cacheItem).value = value
		elem.Value.(*cacheItem).expiration = expiration
		s.lruList.MoveToFront(elem)

		if expiration != nil {
			s.withTTL[key] = true
		} else {
			delete(s.withTTL, key)
		}

		return
	}

	// Add new item
	item := &cacheItem{key: key, value: value, expiration: expiration}
	elem := s.lruList.PushFront(item)
	s.data[key] = elem

	if expiration != nil {
		s.withTTL[key] = true
	}

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
		delete(s.withTTL, item.key)
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

func (s *Store) GetAllKeys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	elem, ok := s.data[key]
	if !ok {
		return "", false
	}

	expir := elem.Value.(*cacheItem).expiration

	if expir != nil {
		if expir.Before(time.Now()) {
			delete(s.data, key)
			s.lruList.Remove(elem)
			return "", false
		}
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
	store := &Store{
		data:    make(map[string]*list.Element),
		lruList: list.New(),
		maxSize: 1000, // Default max size, could be configurable
		withTTL: make(map[string]bool),
	}

	go store.cleanup()

	return store
}

func (s *Store) cleanup() {
	for {
		s.mu.Lock()
		if len(s.withTTL) == 0 {
			s.mu.Unlock()
			time.Sleep(time.Second)
			continue
		}

		start := time.Now()
		for {
			var keySlice []string
			for key := range s.withTTL {
				keySlice = append(keySlice, key)
			}

			count := len(keySlice)
			expired := 0
			sampleSize := min(20, count)
			min := 0
			max := count - 1

			for i := 0; i < sampleSize; i++ {
				elapsed := time.Since(start)
				if elapsed >= 25*time.Millisecond {
					break
				}

				idx := rand.Intn(max-min+1) + min
				key := keySlice[idx]
				elem := s.data[key]

				if elem == nil {
					continue
				}

				expiration := elem.Value.(*cacheItem).expiration
				if expiration != nil {
					isExpired := expiration.Before(time.Now())
					if isExpired {
						expired++
						delete(s.data, key)
						delete(s.withTTL, key)
						s.lruList.Remove(elem)
					}
				}
			}

			lotsOfExpired := float64(expired)/float64(sampleSize) > 0.25
			isTimeLeft := time.Since(start) <= 25*time.Millisecond

			if lotsOfExpired && isTimeLeft {
				continue
			} else {
				break
			}
		}

		s.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}
