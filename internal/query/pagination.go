package query

import (
	"sort"

	"github.com/121watts/reredis/internal/store"
)

type PaginationResult struct {
	Keys       []string `json:"keys"`
	NextCursor string   `json:"next_cursor"`
	HasMore    bool     `json:"has_more"`
}

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
