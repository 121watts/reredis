package cluster

import (
	"testing"
)

func TestClusterManagerLifecycle(t *testing.T) {
	t.Run("single node cluster manager", func(t *testing.T) {
		manager := NewManager("localhost", "6379")
		
		// Should have one node (itself)
		if len(manager.Nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(manager.Nodes))
		}
		
		// Node should not have valid slots initially
		if manager.Node.Slot.Start != -1 || manager.Node.Slot.End != -1 {
			t.Errorf("Expected uninitialized slots [-1, -1], got [%d, %d]", 
				manager.Node.Slot.Start, manager.Node.Slot.End)
		}
		
		// Should be able to find the node by ID
		if manager.Nodes[manager.Node.ID] != manager.Node {
			t.Error("Manager's own node not found in nodes map")
		}
	})
	
	t.Run("two node cluster - no initialization", func(t *testing.T) {
		manager := NewManager("localhost", "6379")
		manager.AddNode("localhost", "6380")
		
		// Should have 2 nodes but no slot distribution yet
		if len(manager.Nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(manager.Nodes))
		}
		
		// All nodes should still have uninitialized slots
		for _, node := range manager.Nodes {
			if node.Slot.Start != -1 || node.Slot.End != -1 {
				t.Errorf("Node %s should have uninitialized slots, got [%d, %d]", 
					node.ID, node.Slot.Start, node.Slot.End)
			}
		}
	})
	
	t.Run("three node cluster - automatic initialization", func(t *testing.T) {
		manager := NewManager("localhost", "6379")
		manager.AddNode("localhost", "6380")
		manager.AddNode("localhost", "6381")
		
		// Should have 3 nodes with slot distribution
		if len(manager.Nodes) != 3 {
			t.Errorf("Expected 3 nodes, got %d", len(manager.Nodes))
		}
		
		// All nodes should have valid slots
		for nodeID, node := range manager.Nodes {
			if node.Slot.Start == -1 || node.Slot.End == -1 {
				t.Errorf("Node %s should have initialized slots, got [%d, %d]", 
					nodeID, node.Slot.Start, node.Slot.End)
			}
			
			if node.Slot.Start > node.Slot.End {
				t.Errorf("Node %s has invalid slot range [%d, %d]", 
					nodeID, node.Slot.Start, node.Slot.End)
			}
		}
		
		// Verify complete slot coverage
		verifySlotCoverage(t, manager)
	})
}

func TestSlotRouting(t *testing.T) {
	manager := NewManager("localhost", "6379")
	manager.AddNode("localhost", "6380")
	manager.AddNode("localhost", "6381")
	
	// Test key routing scenarios
	testKeys := []string{
		"user:1001",
		"session:abc123",
		"cache:temperature",
		"data:metrics:cpu",
		"temp:file:upload",
		"",
		"a",
		"very:long:key:with:many:segments:that:might:hash:differently",
	}
	
	for _, key := range testKeys {
		slot := CalculateSlot(key)
		node := manager.GetNodeForSlots(slot)
		
		if node == nil {
			t.Errorf("No node found for key '%s' (slot %d)", key, slot)
			continue
		}
		
		// Verify the slot is actually within the node's range
		if slot < node.Slot.Start || slot > node.Slot.End {
			t.Errorf("Key '%s' (slot %d) routed to node with range [%d, %d]", 
				key, slot, node.Slot.Start, node.Slot.End)
		}
		
		// Verify routing is consistent
		node2 := manager.GetNodeForSlots(slot)
		if node != node2 {
			t.Errorf("Inconsistent routing for key '%s': got different nodes", key)
		}
	}
}

func TestSlotRangeEdgeCases(t *testing.T) {
	manager := NewManager("localhost", "6379")
	manager.AddNode("localhost", "6380")
	manager.AddNode("localhost", "6381")
	
	// Test edge slot values
	testSlots := []int32{
		0,                // First slot
		SLOT_RANGE - 1,   // Last slot
		SLOT_RANGE / 2,   // Middle slot
	}
	
	for _, slot := range testSlots {
		node := manager.GetNodeForSlots(slot)
		if node == nil {
			t.Errorf("No node found for edge case slot %d", slot)
		}
	}
	
	// Test invalid slots
	invalidSlots := []int32{
		-1,
		SLOT_RANGE,
		SLOT_RANGE + 1000,
	}
	
	for _, slot := range invalidSlots {
		node := manager.GetNodeForSlots(slot)
		if node != nil {
			t.Errorf("Found node for invalid slot %d, expected nil", slot)
		}
	}
}

func TestClusterExpansion(t *testing.T) {
	// Start with single node
	manager := NewManager("localhost", "6379")
	
	// Add nodes one by one and verify behavior
	for i := 1; i < 5; i++ {
		port := 6379 + i
		manager.AddNode("localhost", string(rune('0'+port)))
		
		nodeCount := len(manager.Nodes)
		if nodeCount != i+1 {
			t.Errorf("After adding node %d, expected %d nodes, got %d", 
				i, i+1, nodeCount)
		}
		
		// Only when we have exactly 3 nodes should slot distribution happen
		if nodeCount == 3 {
			// Verify all nodes have valid slots
			for nodeID, node := range manager.Nodes {
				if node.Slot.Start == -1 || node.Slot.End == -1 {
					t.Errorf("Node %s should have slots after cluster init", nodeID)
				}
			}
			
			verifySlotCoverage(t, manager)
		} else if nodeCount < 3 {
			// Nodes < 3 should not have slot distribution
			for nodeID, node := range manager.Nodes {
				if node.Slot.Start != -1 || node.Slot.End != -1 {
					t.Errorf("Node %s should not have slots before cluster init, got [%d, %d]", 
						nodeID, node.Slot.Start, node.Slot.End)
				}
			}
		} else {
			// Nodes > 3: the first 3 should have slots, others should not
			nodesWithSlots := 0
			for _, node := range manager.Nodes {
				if node.Slot.Start != -1 && node.Slot.End != -1 {
					nodesWithSlots++
				}
			}
			if nodesWithSlots != 3 {
				t.Errorf("Expected exactly 3 nodes to have slots when cluster size > 3, got %d", nodesWithSlots)
			}
		}
	}
}

func TestNodeAddressHandling(t *testing.T) {
	manager := NewManager("192.168.1.10", "6379")
	manager.AddNode("192.168.1.11", "6380")
	manager.AddNode("192.168.1.12", "6381")
	
	// Verify each node has correct address info
	addressSet := make(map[string]bool)
	for _, node := range manager.Nodes {
		address := node.Host + ":" + node.Port
		if addressSet[address] {
			t.Errorf("Duplicate address found: %s", address)
		}
		addressSet[address] = true
		
		if node.Host == "" || node.Port == "" {
			t.Errorf("Node %s has empty host or port: %s:%s", 
				node.ID, node.Host, node.Port)
		}
	}
	
	// Should have 3 unique addresses
	if len(addressSet) != 3 {
		t.Errorf("Expected 3 unique addresses, got %d", len(addressSet))
	}
}

func TestConcurrentSlotLookup(t *testing.T) {
	manager := NewManager("localhost", "6379")
	manager.AddNode("localhost", "6380")
	manager.AddNode("localhost", "6381")
	
	// Test concurrent access to slot lookup
	done := make(chan bool, 100)
	
	for i := 0; i < 100; i++ {
		go func(keyNum int) {
			key := string(rune('a' + keyNum%26))
			slot := CalculateSlot(key)
			node := manager.GetNodeForSlots(slot)
			
			if node == nil {
				t.Errorf("Concurrent lookup failed for key %s", key)
			}
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// Helper function to verify complete slot coverage
func verifySlotCoverage(t *testing.T, manager *Manager) {
	t.Helper()
	
	slotOwnership := make(map[int32]*Node)
	
	// Map each slot to its owner (only for nodes with valid slot assignments)
	for _, node := range manager.Nodes {
		// Skip nodes with uninitialized slots
		if node.Slot.Start == -1 || node.Slot.End == -1 {
			continue
		}
		
		for slot := node.Slot.Start; slot <= node.Slot.End; slot++ {
			if existingOwner, exists := slotOwnership[slot]; exists {
				t.Errorf("Slot %d is owned by both %s and %s", 
					slot, existingOwner.ID, node.ID)
			}
			slotOwnership[slot] = node
		}
	}
	
	// Verify all slots are covered (use the constant directly since we're in the same package)
	for slot := int32(0); slot < SLOT_RANGE; slot++ {
		if _, covered := slotOwnership[slot]; !covered {
			t.Errorf("Slot %d is not covered by any node", slot)
		}
	}
	
	// Verify total coverage
	if len(slotOwnership) != int(SLOT_RANGE) {
		t.Errorf("Expected %d slots to be covered, got %d", SLOT_RANGE, len(slotOwnership))
	}
}