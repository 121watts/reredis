package cluster

import (
	"hash/crc32"
	"strings"
)

const SLOT_RANGE int32 = 16_384

func CalculateSlot(key string) int32 {
	// Redis hash tag support: if key contains {tag}, hash only the tag content
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

func GetSlotRange() int32 {
	return SLOT_RANGE
}
