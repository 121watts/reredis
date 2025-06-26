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
	"github.com/121watts/reredis/internal/wal"
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

	// Create WAL writer (placeholder for now)
	walWriter, err := wal.NewWriter("reredis.wal")
	if err != nil {
		return fmt.Errorf("failed to create WAL writer: %w", err)
	}
	defer walWriter.Close()

	// Create command handler
	handler := NewCommandHandler(s, hub, walWriter, cm, logger)

	for {
		conn, err := ln.Accept()

		if err != nil {
			logger.Error("failed to accept connection", "error", err)
			continue
		}

		lw := logger.With("remote_addr", conn.RemoteAddr().String())

		go handleConnection(conn, lw, handler)
	}
}

// commandTable maps Redis command names to their handler functions.
// This enables fast command dispatch and easy extension with new Redis-compatible
// commands while maintaining clean separation of concerns.
func handleCommand(cmd string, parts []string, conn net.Conn, logger *slog.Logger, handler *CommandHandler) {
	// Check for redirect on key-based commands
	if isKeyBasedCommand(cmd) && len(parts) >= 2 {
		key := parts[1]
		if redirectKey := handler.checkSlotOwnership(key); redirectKey != "" {
			handleRedirect(redirectKey, conn, handler)
			return
		}
	}

	switch cmd {
	case "SET":
		handleSetCommand(parts, conn, logger, handler)
	case "GET":
		handleGetCommand(parts, conn, logger, handler)
	case "DEL":
		handleDeleteCommand(parts, conn, logger, handler)
	case "CLUSTER":
		handleClusterCommand(parts, conn, logger, handler)
	default:
		fmt.Fprintf(conn, "-ERR unknown command\r\n")
	}
}

func isKeyBasedCommand(cmd string) bool {
	switch cmd {
	case "SET", "GET", "DEL":
		return true
	default:
		return false
	}
}

// handleConnection processes Redis protocol commands from a single client connection.
// This implements the Redis wire protocol with proper parsing and response formatting,
// enabling compatibility with existing Redis clients and tools.
func handleConnection(conn net.Conn, logger *slog.Logger, handler *CommandHandler) {
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
		handleCommand(cmd, parts, conn, logger, handler)
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

func handleSetCommand(parts []string, conn net.Conn, logger *slog.Logger, handler *CommandHandler) {
	result, err := handler.HandleSet(parts)
	if err != nil {
		fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
		return
	}

	// Send response
	fmt.Fprintf(conn, "+OK\r\n")

	// Handle broadcasting
	if result != nil {
		handler.hub.BroadcastMessage(observer.UpdateMessage{
			Action: result.Action,
			Key:    result.Key,
			Value:  result.Value,
		})

		if result.NeedsStats {
			broadcastClusterStats(handler.hub, handler.store, handler.clusterManager)
		}
	}
}

func handleGetCommand(parts []string, conn net.Conn, logger *slog.Logger, handler *CommandHandler) {
	value, err := handler.HandleGet(parts)
	if err != nil {
		fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
	} else {
		fmt.Fprintf(conn, "%s\r\n", value)
	}
}

func handleDeleteCommand(parts []string, conn net.Conn, logger *slog.Logger, handler *CommandHandler) {
	deleted, result, err := handler.HandleDelete(parts)
	if err != nil {
		fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
		return
	}

	// Send response
	if deleted {
		fmt.Fprintf(conn, ":1\r\n")
	} else {
		fmt.Fprintf(conn, ":0\r\n")
	}

	// Handle broadcasting
	if result != nil {
		handler.hub.BroadcastMessage(observer.UpdateMessage{
			Action: result.Action,
			Key:    result.Key,
		})

		if result.NeedsStats {
			broadcastClusterStats(handler.hub, handler.store, handler.clusterManager)
		}
	}
}

func handleClusterCommand(parts []string, conn net.Conn, logger *slog.Logger, handler *CommandHandler) {
	err := handler.HandleCluster(parts)
	if err != nil {
		fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
	} else {
		fmt.Fprintf(conn, "+OK\r\n")
	}
}

func handleRedirect(key string, conn net.Conn, handler *CommandHandler) {
	if handler.clusterManager == nil {
		return
	}

	slot := cluster.CalculateSlot(key)
	ownerNode := handler.clusterManager.GetNodeForSlots(slot)
	if ownerNode != nil {
		fmt.Fprintf(conn, "-MOVED %d %s:%s\r\n", slot, ownerNode.Host, ownerNode.Port)
	} else {
		fmt.Fprintf(conn, "-ERR no node found for slot %d\r\n", slot)
	}
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
