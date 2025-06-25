// Package cluster implements Redis-compatible clustering functionality.
// This provides distributed key-value storage across multiple nodes using
// hash slots for consistent data partitioning.
package cluster

import (
	"hash/crc32"
	"strings"
)

// SLOT_RANGE defines the total number of hash slots in the cluster.
// Redis uses 16384 slots to balance memory usage with distribution granularity,
// allowing fine-grained data migration while keeping slot tables manageable.
const SLOT_RANGE int32 = 16_384

// CalculateSlot determines which hash slot a key belongs to for cluster routing.
// This enables consistent data distribution across nodes and supports Redis
// hash tags for keeping related keys on the same node when needed.
func CalculateSlot(key string) int32 {
	// Redis hash tag support: if key contains {tag}, hash only the tag content
	// This allows developers to control key placement by ensuring keys with
	// the same tag go to the same node, enabling multi-key operations
	hashKey := key
	if start := strings.Index(key, "{"); start != -1 {
		if end := strings.Index(key[start+1:], "}"); end != -1 {
			hashKey = key[start+1 : start+1+end]
		}
	}
	
	keyBytes := []byte(hashKey)
	hashResult := crc32.ChecksumIEEE(keyBytes)

	slot := int32(hashResult % uint32(SLOT_RANGE))

	return slot
}

// GetSlotRange returns the total number of hash slots available.
// This provides external access to the slot range constant while maintaining
// encapsulation and allowing future changes to the slot range implementation.
func GetSlotRange() int32 {
	return SLOT_RANGE
}
