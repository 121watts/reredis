package cluster

import (
	"hash/crc32"
)

const SLOT_RANGE int32 = 16_384

func CalculateSlot(key string) int32 {
	keyBytes := []byte(key)
	hashResult := int32(crc32.ChecksumIEEE(keyBytes))

	slot := hashResult % SLOT_RANGE

	return slot
}
