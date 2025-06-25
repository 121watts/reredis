package cluster_test

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/121watts/reredis/internal/cluster"
	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
)

func startClusterTestServer(t *testing.T, host, port string) (*cluster.Manager, string) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := observer.NewHub(logger)
	go hub.Run()

	s := store.NewStore()
	clusterManager := cluster.NewManager(host, port)

	ln, err := net.Listen("tcp", host+":0") // Use dynamic port
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	go func() {
		if err := server.StartWithListener(ln, s, logger, hub, clusterManager); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	// Wait for server to start
	const maxRetries = 10
	const retryDelay = 100 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return clusterManager, addr
		}
		time.Sleep(retryDelay)
	}

	t.Fatalf("server did not start in time")
	return nil, ""
}

func sendClusterCommand(t *testing.T, conn net.Conn, cmd string) string {
	t.Helper()

	_, err := fmt.Fprintf(conn, "%s\r\n", cmd)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	return resp
}

func TestClusterManagerIntegration(t *testing.T) {
	t.Run("single node key operations", func(t *testing.T) {
		manager, addr := startClusterTestServer(t, "localhost", "6379")
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("could not connect: %v", err)
		}
		defer conn.Close()

		// Test key operations on single node (no cluster yet)
		testKeys := []struct {
			key   string
			value string
		}{
			{"user:1001", "john"},
			{"session:abc", "active"},
			{"cache:temp", "25C"},
		}

		for _, test := range testKeys {
			// Calculate which slot this key would belong to
			slot := cluster.CalculateSlot(test.key)
			node := manager.GetNodeForSlots(slot)

			// Since we only have one node, it should handle all slots
			if node == nil {
				t.Errorf("No node found for key %s (slot %d)", test.key, slot)
				continue
			}

			// Set the key
			setResp := sendClusterCommand(t, conn, fmt.Sprintf("SET %s %s", test.key, test.value))
			if !strings.Contains(setResp, "OK") {
				t.Errorf("SET failed for key %s: %s", test.key, setResp)
			}

			// Get the key
			getResp := sendClusterCommand(t, conn, fmt.Sprintf("GET %s", test.key))
			expectedResp := test.value + "\r\n"
			if getResp != expectedResp {
				t.Errorf("GET %s: expected %q, got %q", test.key, expectedResp, getResp)
			}
		}
	})

	t.Run("three node cluster formation", func(t *testing.T) {
		manager := cluster.NewManager("localhost", "6379")

		// Simulate adding nodes as they would join the cluster
		manager.AddNode("localhost", "6380")
		if len(manager.Nodes) != 2 {
			t.Errorf("Expected 2 nodes after first addition, got %d", len(manager.Nodes))
		}

		// Verify no slot assignment yet
		for _, node := range manager.Nodes {
			if node.Slot.Start != -1 || node.Slot.End != -1 {
				t.Errorf("Node should not have slots before 3-node threshold")
			}
		}

		// Add third node - should trigger cluster initialization
		manager.AddNode("localhost", "6381")
		if len(manager.Nodes) != 3 {
			t.Errorf("Expected 3 nodes after second addition, got %d", len(manager.Nodes))
		}

		// Verify slot assignment happened
		totalSlots := int32(0)
		for nodeID, node := range manager.Nodes {
			if node.Slot.Start == -1 || node.Slot.End == -1 {
				t.Errorf("Node %s should have slot assignment after cluster init", nodeID)
			}
			totalSlots += (node.Slot.End - node.Slot.Start + 1)
		}

		if totalSlots != cluster.GetSlotRange() {
			t.Errorf("Total slots assigned (%d) != SLOT_RANGE (%d)", totalSlots, cluster.GetSlotRange())
		}
	})

	t.Run("slot lookup performance", func(t *testing.T) {
		manager := cluster.NewManager("localhost", "6379")
		manager.AddNode("localhost", "6380")
		manager.AddNode("localhost", "6381")

		// Test large number of key lookups
		const numLookups = 10000
		keys := make([]string, numLookups)
		for i := 0; i < numLookups; i++ {
			keys[i] = fmt.Sprintf("key:%d", i)
		}

		start := time.Now()
		for _, key := range keys {
			slot := cluster.CalculateSlot(key)
			node := manager.GetNodeForSlots(slot)
			if node == nil {
				t.Errorf("Failed to find node for key %s", key)
			}
		}
		elapsed := time.Since(start)

		// Should be very fast - let's say under 100ms for 10k lookups
		if elapsed > 100*time.Millisecond {
			t.Errorf("Slot lookups too slow: %v for %d lookups", elapsed, numLookups)
		}

		avgPerLookup := elapsed.Nanoseconds() / int64(numLookups)
		t.Logf("Average lookup time: %d ns", avgPerLookup)
	})
}

func TestSlotDistributionQuality(t *testing.T) {
	manager := cluster.NewManager("localhost", "6379")
	manager.AddNode("localhost", "6380")
	manager.AddNode("localhost", "6381")

	// Generate many keys and see how they distribute
	const numKeys = 10000
	nodeKeyCounts := make(map[string]int)

	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("user:%d:profile", i)
		slot := cluster.CalculateSlot(key)
		node := manager.GetNodeForSlots(slot)

		if node == nil {
			t.Errorf("No node found for key %s", key)
			continue
		}

		nodeKeyCounts[node.ID]++
	}

	// Check that distribution is reasonably balanced
	// With 3 nodes, we expect roughly 3333 keys per node
	expectedPerNode := numKeys / 3
	tolerance := expectedPerNode / 5 // 20% tolerance

	for nodeID, count := range nodeKeyCounts {
		if count < expectedPerNode-tolerance || count > expectedPerNode+tolerance {
			t.Errorf("Node %s has unbalanced key distribution: %d keys (expected ~%d)",
				nodeID, count, expectedPerNode)
		}
	}

	t.Logf("Key distribution across nodes: %v", nodeKeyCounts)
}

func TestHashtagSupport(t *testing.T) {
	// Test Redis hash tag functionality (keys with same hash tag go to same slot)
	testCases := []struct {
		key1             string
		key2             string
		shouldBeSameSlot bool
	}{
		{"user{123}:profile", "user{123}:settings", true},
		{"session{abc}:data", "session{abc}:meta", true},
		{"user{123}:profile", "user{456}:profile", false},
		{"key1", "key2", false},
		{"{shared}:data1", "{shared}:data2", true},
		{"prefix{tag}suffix", "other{tag}end", true},
	}

	for _, tc := range testCases {
		slot1 := cluster.CalculateSlot(tc.key1)
		slot2 := cluster.CalculateSlot(tc.key2)

		sameSlot := slot1 == slot2
		if sameSlot != tc.shouldBeSameSlot {
			t.Errorf("Keys %s and %s: expected same slot = %v, got slots %d and %d",
				tc.key1, tc.key2, tc.shouldBeSameSlot, slot1, slot2)
		}
	}
}

func TestClusterManagerEdgeCases(t *testing.T) {
	t.Run("empty host and port", func(t *testing.T) {
		manager := cluster.NewManager("", "")
		if manager.Node.Host != "" || manager.Node.Port != "" {
			t.Error("Manager should handle empty host/port")
		}
	})

	t.Run("duplicate node addition", func(t *testing.T) {
		manager := cluster.NewManager("localhost", "6379")

		initialCount := len(manager.Nodes)
		manager.AddNode("localhost", "6379") // Same as manager's own node

		// This creates a new node ID even with same address
		if len(manager.Nodes) != initialCount+1 {
			t.Errorf("Expected node count to increase even with duplicate address")
		}
	})

	t.Run("large cluster", func(t *testing.T) {
		manager := cluster.NewManager("localhost", "6379")

		// Add many nodes to test scalability
		for i := 1; i < 100; i++ {
			manager.AddNode("localhost", fmt.Sprintf("%d", 6379+i))
		}

		if len(manager.Nodes) != 100 {
			t.Errorf("Expected 100 nodes, got %d", len(manager.Nodes))
		}

		// Only first 3 nodes should have slot assignments
		nodesWithSlots := 0
		for _, node := range manager.Nodes {
			if node.Slot.Start != -1 && node.Slot.End != -1 {
				nodesWithSlots++
			}
		}

		if nodesWithSlots != 3 {
			t.Errorf("Expected 3 nodes with slots, got %d", nodesWithSlots)
		}
	})
}
