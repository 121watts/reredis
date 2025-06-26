// Package store provides a thread-safe in-memory key-value store with LRU eviction and TTL support.
// This combines Redis-like semantics with automatic memory management, enabling predictable
// performance for caching and session storage in distributed applications.
package store

import (
	"container/list"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// cacheItem represents a single key-value pair with optional expiration.
// This encapsulates the data needed for both LRU tracking and TTL management,
// enabling efficient memory usage and automatic cleanup.
type cacheItem struct {
	key        string
	value      string
	expiration *time.Time // nil means no expiration
}

// Store provides a thread-safe LRU cache with TTL support and automatic cleanup.
// This balances memory usage with performance by evicting old data and expired keys,
// making it suitable for high-throughput applications with predictable memory requirements.
type Store struct {
	data    map[string]*list.Element // Fast key lookup to list elements for O(1) access
	lruList *list.List               // Doubly linked list for efficient LRU operations
	maxSize int                      // Prevents unbounded memory growth
	mu      sync.Mutex               // Ensures thread safety for concurrent access
	withTTL map[string]bool          // Tracks keys with TTL for efficient cleanup sampling
}

// Set stores a key-value pair without expiration.
// This provides permanent storage until evicted by LRU policy, making it suitable
// for configuration data and long-lived application state.
func (s *Store) Set(key, value string) {
	s.setInternal(key, value, nil)
}

// SetWithTTL stores a key-value pair with automatic expiration after the specified duration.
// This enables temporary data storage for sessions, caches, and rate limiting,
// reducing memory usage and providing automatic cleanup.
func (s *Store) SetWithTTL(key, value string, ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	s.setInternal(key, value, &expiration)
}

// setInternal handles the core storage logic for both permanent and TTL-based keys.
// This centralizes the LRU management and TTL tracking, ensuring consistent behavior
// and optimal performance across different storage scenarios.
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

// evictLRU removes the least recently used item to maintain memory limits.
// This prevents unbounded memory growth while preserving the most valuable data,
// ensuring predictable performance even under heavy load.
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

// GetAll returns a snapshot of all key-value pairs currently in the store.
// This enables bulk operations and state synchronization, providing a consistent
// view of the data at a specific point in time for debugging and replication.
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

// GetAllKeys returns a sorted list of all keys currently in the store.
// This supports administrative operations and key enumeration for applications
// that need to iterate over stored data in a predictable order.
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

// Get retrieves a value by key and handles expiration automatically.
// This provides lazy expiration to avoid memory leaks while maintaining LRU
// ordering for optimal cache performance and accurate hit ratios.
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

// Delete removes a key-value pair from the store if it exists.
// This supports data cleanup and cache invalidation, returning whether
// the key was actually present to enable proper application logic.
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	elem, ok := s.data[key]
	if ok {
		delete(s.data, key)
		delete(s.withTTL, key)
		s.lruList.Remove(elem)
	}

	return ok
}

// NewStore creates a new thread-safe store with automatic cleanup and LRU eviction.
// This initializes the data structures and background processes needed for efficient
// memory management and TTL expiration in high-concurrency environments.
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

// cleanup runs a background process to actively expire TTL keys.
// This implements Redis-like active expiration to prevent memory buildup,
// using probabilistic sampling to balance CPU usage with memory efficiency.
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

// GetTotalByteSize calculates the total byte size of all stored data.
// This provides storage usage statistics for monitoring and capacity planning,
// including both keys and values in the calculation.
func (s *Store) GetTotalByteSize() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var totalSize int64
	
	for _, elem := range s.data {
		item := elem.Value.(*cacheItem)
		// Count both key and value bytes (UTF-8 encoded)
		totalSize += int64(len(item.key) + len(item.value))
	}
	
	return totalSize
}
