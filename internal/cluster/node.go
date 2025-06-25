package cluster

import (
	"fmt"
)

type Slot struct {
	Start int32
	End   int32
}

type Node struct {
	ID   string
	Slot Slot
	Host string
	Port string
}

func BuildNodes(n int32) []*Node {
	var rangeStart int32 = 0
	rangeEnd := SLOT_RANGE / n
	nodes := []*Node{
		{
			ID:   "0",
			Slot: Slot{Start: rangeStart, End: rangeEnd},
		},
	}

	for i := int32(1); i < n; i++ {
		prevEnd := rangeEnd
		rangeEnd += SLOT_RANGE / n
		nodes = append(nodes, &Node{
			ID:   fmt.Sprintf("%d", i),
			Slot: Slot{Start: prevEnd, End: rangeEnd},
		})
	}

	return nodes
}
