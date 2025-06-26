package cluster

import (
	"crypto/rand"
	"fmt"
	"sort"
)

// Manager coordinates cluster operations and maintains the distributed state.
// This centralizes cluster topology management, enabling consistent routing
// decisions and cluster lifecycle operations across all nodes.
type Manager struct {
	Nodes map[string]*Node // All known nodes in the cluster for routing decisions
	Node  *Node            // This server's node identity within the cluster
}

// NewManager creates a cluster manager for a single node server.
// This establishes the initial cluster state and node identity, allowing
// the server to operate standalone or join a larger cluster later.
func NewManager(host, port string) *Manager {
	nodeId := generateNodeID()
	node := &Node{
		ID:   nodeId,
		Slot: Slot{Start: -1, End: -1}, // Uninitialized slots until cluster forms
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

// AddNode registers a new node in the cluster topology.
// This enables cluster expansion and triggers automatic slot distribution
// once enough nodes join to form a viable cluster.
func (m *Manager) AddNode(host, port string) {
	newNodeId := generateNodeID()
	m.Nodes[newNodeId] = &Node{
		ID:   newNodeId,
		Slot: Slot{Start: -1, End: -1}, // Will be assigned during initialization
		Host: host,
		Port: port,
	}

	// Initialize cluster when we have 3 nodes
	// Redis requires minimum 3 nodes for proper cluster operation and failover
	if len(m.Nodes) == 3 {
		m.InitializeCluster()
	}
}

// InitializeCluster distributes hash slots evenly across available nodes.
// This creates a balanced data distribution and enables the cluster to handle
// client requests with predictable performance characteristics.
func (m *Manager) InitializeCluster() {
	// Get all node IDs and sort them for consistent slot assignment
	// Sorting ensures deterministic slot distribution across cluster restarts
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

		// Last node gets any remaining slots to ensure complete coverage
		if i == len(nodeIDs)-1 {
			end = SLOT_RANGE - 1
		}

		m.Nodes[nodeID].Slot = Slot{Start: start, End: end}
	}
}

// GetNodeForSlots finds which node is responsible for a given hash slot.
// This enables request routing to the correct node and supports MOVED
// redirections when clients contact the wrong node for a key.
func (m *Manager) GetNodeForSlots(slot int32) *Node {
	// Defensive check: if manager is nil, return nil
	if m == nil {
		return nil
	}
	
	// If cluster is not initialized (< 3 nodes), return the current node for any slot
	// This allows single nodes to handle all operations during development/testing
	if len(m.Nodes) < 3 {
		return m.Node
	}

	for _, node := range m.Nodes {
		// Defensive check: skip nil nodes
		if node == nil {
			continue
		}
		if slot >= node.Slot.Start && slot <= node.Slot.End {
			return node
		}
	}

	// If no node found, return current node as fallback to prevent nil pointer panics
	// This ensures cluster operations continue even in edge cases
	return m.Node
}

// generateNodeID creates a unique identifier for cluster nodes.
// This ensures each node can be distinctly identified during cluster
// operations and maintains consistency across network communications.
func generateNodeID() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// UpdateKeyCount updates the key count for the current node.
// This maintains accurate statistics for cluster monitoring and load balancing decisions.
func (m *Manager) UpdateKeyCount(count int) {
	if m.Node != nil {
		m.Node.KeyCount = count
		// Also update in the nodes map
		if node, exists := m.Nodes[m.Node.ID]; exists {
			node.KeyCount = count
		}
	}
}

// UpdateByteSize updates the byte size for the current node.
// This maintains accurate storage statistics for cluster monitoring and capacity planning.
func (m *Manager) UpdateByteSize(size int64) {
	if m.Node != nil {
		m.Node.ByteSize = size
		// Also update in the nodes map
		if node, exists := m.Nodes[m.Node.ID]; exists {
			node.ByteSize = size
		}
	}
}

// IncrementKeyCount increments the key count for the current node.
// This is called when a key is added to maintain real-time statistics.
func (m *Manager) IncrementKeyCount() {
	if m.Node != nil {
		m.Node.KeyCount++
		// Also update in the nodes map
		if node, exists := m.Nodes[m.Node.ID]; exists {
			node.KeyCount++
		}
	}
}

// DecrementKeyCount decrements the key count for the current node.
// This is called when a key is deleted to maintain real-time statistics.
func (m *Manager) DecrementKeyCount() {
	if m.Node != nil && m.Node.KeyCount > 0 {
		m.Node.KeyCount--
		// Also update in the nodes map
		if node, exists := m.Nodes[m.Node.ID]; exists && node.KeyCount > 0 {
			node.KeyCount--
		}
	}
}

// AddByteSize adds bytes to the current node's storage count.
// This is called when a key is added to track storage usage.
func (m *Manager) AddByteSize(keySize, valueSize int) {
	if m.Node != nil {
		bytes := int64(keySize + valueSize)
		m.Node.ByteSize += bytes
		// Also update in the nodes map
		if node, exists := m.Nodes[m.Node.ID]; exists {
			node.ByteSize += bytes
		}
	}
}

// SubtractByteSize subtracts bytes from the current node's storage count.
// This is called when a key is deleted to track storage usage.
func (m *Manager) SubtractByteSize(keySize, valueSize int) {
	if m.Node != nil {
		bytes := int64(keySize + valueSize)
		if m.Node.ByteSize >= bytes {
			m.Node.ByteSize -= bytes
		} else {
			m.Node.ByteSize = 0
		}
		// Also update in the nodes map
		if node, exists := m.Nodes[m.Node.ID]; exists {
			if node.ByteSize >= bytes {
				node.ByteSize -= bytes
			} else {
				node.ByteSize = 0
			}
		}
	}
}
