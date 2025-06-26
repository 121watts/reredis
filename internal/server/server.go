package server

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/121watts/reredis/internal/cluster"
	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/store"
)

// Start launches the Redis-compatible TCP server on the specified address.
// This provides the main Redis protocol interface, enabling existing Redis clients
// to connect and operate with full compatibility for commands like SET, GET, DEL.
func Start(address string, s *store.Store, logger *slog.Logger, hub *observer.Hub, cm *cluster.Manager) error {
	ln, err := net.Listen("tcp", address)

	if err != nil {
		return fmt.Errorf("failed to bind: %w", err)
	}

	return StartWithListener(ln, s, logger, hub, cm)
}

// StartWithListener runs the TCP server using an existing network listener.
// This enables testing with dynamic ports and supports advanced deployment
// scenarios where the listener is managed externally.
func StartWithListener(ln net.Listener, s *store.Store, logger *slog.Logger, hub *observer.Hub, cm *cluster.Manager) error {
	defer ln.Close()
	logger.Info("listening on port", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()

		if err != nil {
			logger.Error("failed to accept connection", "error", err)
			continue
		}

		lw := logger.With("remote_addr", conn.RemoteAddr().String())

		go handleConnection(conn, s, lw, hub, cm)
	}
}

// command defines the signature for Redis command handlers.
// This standardizes command processing and enables extensible command registration
// while maintaining consistent access to store, networking, and clustering components.
type command func(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager)

// commandTable maps Redis command names to their handler functions.
// This enables fast command dispatch and easy extension with new Redis-compatible
// commands while maintaining clean separation of concerns.
var commandTable = map[string]command{
	"SET":     handleSet,
	"GET":     handleGet,
	"DEL":     handleDelete,
	"CLUSTER": handleCluster,
}

// handleConnection processes Redis protocol commands from a single client connection.
// This implements the Redis wire protocol with proper parsing and response formatting,
// enabling compatibility with existing Redis clients and tools.
func handleConnection(conn net.Conn, s *store.Store, logger *slog.Logger, hub *observer.Hub, cm *cluster.Manager) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		parts, err := parseRedisCommand(line)

		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
			continue
		}

		if len(parts) == 0 {
			fmt.Fprintf(conn, "-ERR empty command\r\n")
			continue
		}

		cmd := strings.ToUpper(parts[0])

		handler, ok := commandTable[cmd]
		if !ok {
			fmt.Fprintf(conn, "-ERR unknown command\r\n")
			continue
		}
		handler(parts, s, conn, hub, logger, cm)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("error reading from connection", "error", err)
	}
}

// parseRedisCommand parses a Redis command line, handling quoted strings with spaces.
// This enables proper Redis protocol compatibility by correctly parsing arguments
// that contain spaces when enclosed in double quotes.
func parseRedisCommand(line string) ([]string, error) {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escaped := false
	
	runes := []rune(line)
	
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		
		if r == '\\' {
			escaped = true
			continue
		}
		
		if r == '"' {
			inQuotes = !inQuotes
			continue
		}
		
		if !inQuotes && (r == ' ' || r == '\t') {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			// Skip additional whitespace
			for i+1 < len(runes) && (runes[i+1] == ' ' || runes[i+1] == '\t') {
				i++
			}
			continue
		}
		
		current.WriteRune(r)
	}
	
	if inQuotes {
		return nil, fmt.Errorf("unclosed quoted string")
	}
	
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	return parts, nil
}

// checkSlotOwnership verifies if the current node should handle a key's operation.
// This enforces cluster data partitioning by redirecting clients to the correct
// node when they contact the wrong server, maintaining data consistency across the cluster.
func checkSlotOwnership(key string, cm *cluster.Manager, conn net.Conn) bool {
	// Defensive check: if cluster manager is nil, allow all operations
	if cm == nil {
		return true
	}
	
	// If cluster is not initialized (< 3 nodes), current node handles all slots
	// This allows single nodes to operate normally during development and testing
	if len(cm.Nodes) < 3 {
		return true
	}
	
	// Defensive check: if current node is nil, allow all operations
	if cm.Node == nil {
		return true
	}
	
	slot := cluster.CalculateSlot(key)
	node := cm.Node

	if slot < node.Slot.Start || slot > node.Slot.End {
		ownerNode := cm.GetNodeForSlots(slot)
		if ownerNode != nil {
			// Send MOVED response to redirect client to the correct node
			// This follows Redis cluster protocol for client-side routing
			fmt.Fprintf(conn, "-MOVED %d %s:%s\r\n", slot, ownerNode.Host, ownerNode.Port)
		} else {
			fmt.Fprintf(conn, "-ERR no node found for slot %d\r\n", slot)
		}
		return false
	}
	return true
}

// broadcastClusterStats creates and broadcasts current cluster statistics to all WebSocket clients.
// This provides real-time monitoring updates whenever keys are added or removed from the cluster.
func broadcastClusterStats(hub *observer.Hub, s *store.Store, cm *cluster.Manager) {
	nodes := make([]observer.ClusterNodeStats, 0, len(cm.Nodes))
	totalKeys := 0
	
	for _, node := range cm.Nodes {
		var keyCount int
		var byteSize int64
		
		// For the current node, get the actual stats from the store
		if node.ID == cm.Node.ID {
			keyCount = len(s.GetAll())
			byteSize = s.GetTotalByteSize()
		} else {
			// For other nodes, use the cached values (updated during CLUSTER_INFO)
			keyCount = node.KeyCount
			byteSize = node.ByteSize
		}
		
		totalKeys += keyCount
		
		nodes = append(nodes, observer.ClusterNodeStats{
			ID:        node.ID,
			Host:      node.Host,
			Port:      node.Port,
			SlotStart: node.Slot.Start,
			SlotEnd:   node.Slot.End,
			KeyCount:  keyCount,
			ByteSize:  byteSize,
		})
	}
	
	stats := observer.ClusterStatsMessage{
		Action:        "cluster_stats",
		Nodes:         nodes,
		CurrentNodeID: cm.Node.ID,
		TotalSlots:    int(cluster.SLOT_RANGE),
		ClusterSize:   len(cm.Nodes),
		TotalKeys:     totalKeys,
	}
	
	hub.BroadcastClusterStats(stats)
}

// handleSet processes Redis SET commands with cluster-aware routing.
// This stores key-value pairs while enforcing cluster data partitioning
// and broadcasting changes to WebSocket clients for real-time updates.
func handleSet(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, _ *slog.Logger, cm *cluster.Manager) {
	const expectedParts = 3
	if len(parts) != expectedParts {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'SET'\r\n")
		return
	}

	k, v := parts[1], parts[2]

	if !checkSlotOwnership(k, cm, conn) {
		return
	}

	// Check if this is a new key to update cluster statistics
	oldValue, existsBefore := s.Get(k)
	
	s.Set(k, v)
	
	// Update cluster statistics
	if cm != nil {
		if !existsBefore {
			// New key: increment count and add byte size
			cm.IncrementKeyCount()
			cm.AddByteSize(len(k), len(v))
		} else {
			// Existing key: update byte size (subtract old, add new)
			cm.SubtractByteSize(len(k), len(oldValue))
			cm.AddByteSize(len(k), len(v))
		}
	}
	
	hub.BroadcastMessage(observer.UpdateMessage{
		Action: "set", Key: k, Value: v,
	})
	
	// Broadcast cluster stats update if this is part of a cluster
	if cm != nil && len(cm.Nodes) > 1 {
		broadcastClusterStats(hub, s, cm)
	}
	
	fmt.Fprintf(conn, "+OK\r\n")
}

// handleGet retrieves values for Redis GET commands with cluster routing.
// This enforces data locality by redirecting requests to the correct node
// while providing lazy TTL expiration for accurate data access.
func handleGet(parts []string, s *store.Store, conn net.Conn, _ *observer.Hub, _ *slog.Logger, cm *cluster.Manager) {
	const expectedParts = 2

	if len(parts) != expectedParts {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'GET'\r\n")
		return
	}

	k := parts[1]

	if !checkSlotOwnership(k, cm, conn) {
		return
	}

	v, ok := s.Get(k)

	if !ok {
		fmt.Fprintf(conn, "-ERR key not found\r\n")
		return
	}

	fmt.Fprintf(conn, "%s\r\n", v)
}

// handleDelete processes Redis DEL commands with cluster coordination.
// This removes keys from the appropriate node while broadcasting deletion
// events to WebSocket clients for maintaining data consistency across interfaces.
func handleDelete(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, _ *slog.Logger, cm *cluster.Manager) {
	const expectedParts = 2

	if len(parts) != expectedParts {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'DEL'\r\n")
		return
	}

	k := parts[1]

	if !checkSlotOwnership(k, cm, conn) {
		return
	}

	// Get the value before deletion to track byte size
	oldValue, existed := s.Get(k)
	ok := s.Delete(k)

	if ok {
		// Update cluster statistics when key is deleted
		if cm != nil {
			cm.DecrementKeyCount()
			if existed {
				cm.SubtractByteSize(len(k), len(oldValue))
			}
		}
		
		fmt.Fprintf(conn, ":1\r\n")
		hub.BroadcastMessage(observer.UpdateMessage{
			Action: "del", Key: k,
		})
		
		// Broadcast cluster stats update if this is part of a cluster
		if cm != nil && len(cm.Nodes) > 1 {
			broadcastClusterStats(hub, s, cm)
		}
	} else {
		fmt.Fprintf(conn, ":0\r\n")
	}
}

// handleCluster processes CLUSTER command and its subcommands.
// This provides the Redis cluster management interface, allowing clients
// to discover topology, manage nodes, and monitor cluster health.
func handleCluster(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager) {
	if len(parts) < 2 {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'CLUSTER'\r\n")
		return
	}

	subcommand := strings.ToUpper(parts[1])

	switch subcommand {
	case "MEET":
		handleClusterMeet(parts, s, conn, hub, logger, cm)
	case "NODES":
		handleClusterNodes(parts, s, conn, hub, logger, cm)
	case "INFO":
		handleClusterInfo(parts, s, conn, hub, logger, cm)
	default:
		fmt.Fprintf(conn, "-ERR unknown cluster subcommand '%s'\r\n", subcommand)
	}
}

// handleClusterMeet adds a new node to the cluster topology.
// This implements the Redis CLUSTER MEET command, enabling dynamic cluster
// expansion by allowing existing nodes to introduce new members to the cluster.
func handleClusterMeet(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager) {
	if len(parts) != 4 {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'CLUSTER MEET'\r\n")
		return
	}

	host := parts[2]
	port := parts[3]

	logger.Info("cluster meet requested", "ip", host, "port", port)

	// Add a new node to the cluster (simplified - normally you'd connect and exchange info)
	// This enables cluster discovery and growth by building the node topology
	cm.AddNode(host, port)

	logger.Info("node added to cluster", "total-nodes", len(cm.Nodes))

	fmt.Fprintf(conn, "+OK\r\n")
}

// handleClusterNodes returns information about all known cluster nodes.
// This supports client-side routing by providing the current cluster topology,
// allowing smart clients to cache node locations and route requests efficiently.
func handleClusterNodes(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager) {
	// TODO: Return cluster topology information
	fmt.Fprintf(conn, "+TODO: cluster nodes not implemented\r\n")
}

// handleClusterInfo provides cluster health and status information.
// This enables monitoring and debugging by exposing cluster state, slot coverage,
// and operational metrics to administrators and monitoring tools.
func handleClusterInfo(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager) {
	// TODO: Return cluster status information
	fmt.Fprintf(conn, "+TODO: cluster info not implemented\r\n")
}
