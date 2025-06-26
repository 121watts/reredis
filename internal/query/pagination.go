// Package query provides efficient data access patterns for large datasets.
// This implements cursor-based pagination to enable scalable key enumeration
// without loading entire datasets into memory, supporting admin tools and debugging.
package query

import (
	"sort"

	"github.com/121watts/reredis/internal/store"
)

// PaginationResult represents a page of keys with navigation metadata.
// This enables clients to efficiently iterate through large key sets
// while maintaining stateless operation and predictable memory usage.
type PaginationResult struct {
	Keys       []string `json:"keys"`       // Keys in the current page
	NextCursor string   `json:"next_cursor"` // Token for the next page
	HasMore    bool     `json:"has_more"`    // Whether more data exists
}

// HandleCursorPagination provides efficient key listing with cursor-based navigation.
// This enables administrative tools to browse large datasets without overwhelming
// server memory or network bandwidth, supporting scalable operations.
func HandleCursorPagination(s *store.Store, cursor string, limit int) PaginationResult {
	allKeys := s.GetAllKeys()
	sort.Strings(allKeys)

	var cidx int

	if cursor == "" {
		cidx = 0
	} else {
		cidx = sort.SearchStrings(allKeys, cursor)
		cidx++
	}

	keys := make([]string, 0, limit)

	for i := cidx; i < cidx+limit && i < len(allKeys); i++ {
		k := allKeys[i]
		keys = append(keys, k)
	}

	endIdx := cidx + limit
	nextCursor := ""
	hasMore := endIdx < len(allKeys)

	if hasMore {
		nextCursor = allKeys[endIdx-1]
	}

	return PaginationResult{Keys: keys, NextCursor: nextCursor, HasMore: hasMore}
}
