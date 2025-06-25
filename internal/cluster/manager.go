package cluster

import (
	"crypto/rand"
	"fmt"
	"sort"
)

type Manager struct {
	Nodes map[string]*Node
	Node  *Node
}

func NewManager(host, port string) *Manager {
	nodeId := generateNodeID()
	node := &Node{
		ID:   nodeId,
		Slot: Slot{Start: -1, End: -1},
		Host: host,
		Port: port,
	}

	nodes := map[string]*Node{}
	nodes[nodeId] = node

	m := &Manager{
		Nodes: nodes,
		Node:  node,
	}

	return m
}

func (m *Manager) AddNode(host, port string) {
	newNodeId := generateNodeID()
	m.Nodes[newNodeId] = &Node{
		ID:   newNodeId,
		Slot: Slot{Start: -1, End: -1},
		Host: host,
		Port: port,
	}

	// Initialize cluster when we have 3 nodes
	if len(m.Nodes) == 3 {
		m.InitializeCluster()
	}
}

func (m *Manager) InitializeCluster() {
	// Get all node IDs and sort them for consistent slot assignment
	nodeIDs := make([]string, 0, len(m.Nodes))
	for nodeID := range m.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)

	// Distribute slots evenly among nodes
	slotsPerNode := SLOT_RANGE / int32(len(nodeIDs))

	for i, nodeID := range nodeIDs {
		start := int32(i) * slotsPerNode
		end := start + slotsPerNode - 1

		// Last node gets any remaining slots
		if i == len(nodeIDs)-1 {
			end = SLOT_RANGE - 1
		}

		m.Nodes[nodeID].Slot = Slot{Start: start, End: end}
	}
}

func (m *Manager) GetNodeForSlots(slot int32) *Node {
	for _, node := range m.Nodes {
		if slot >= node.Slot.Start && slot <= node.Slot.End {
			return node
		}
	}

	return nil
}

func generateNodeID() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
