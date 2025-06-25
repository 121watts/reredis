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

func Start(address string, s *store.Store, logger *slog.Logger, hub *observer.Hub, cm *cluster.Manager) error {
	ln, err := net.Listen("tcp", address)

	if err != nil {
		return fmt.Errorf("failed to bind: %w", err)
	}

	return StartWithListener(ln, s, logger, hub, cm)
}

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

type command func(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager)

var commandTable = map[string]command{
	"SET":     handleSet,
	"GET":     handleGet,
	"DEL":     handleDelete,
	"CLUSTER": handleCluster,
}

func handleConnection(conn net.Conn, s *store.Store, logger *slog.Logger, hub *observer.Hub, cm *cluster.Manager) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

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

func checkSlotOwnership(key string, cm *cluster.Manager, conn net.Conn) bool {
	slot := cluster.CalculateSlot(key)
	node := cm.Node

	if slot < node.Slot.Start || slot > node.Slot.End {
		ownerNode := cm.GetNodeForSlots(slot)
		fmt.Fprintf(conn, "-MOVED %d %s:%s\r\n", slot, ownerNode.Host, ownerNode.Port)
		return false
	}
	return true
}

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

	s.Set(k, v)
	hub.BroadcastMessage(observer.UpdateMessage{
		Action: "set", Key: k, Value: v,
	})
	fmt.Fprintf(conn, "+OK\r\n")
}

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

	ok := s.Delete(k)

	if ok {
		fmt.Fprintf(conn, ":1\r\n")
		hub.BroadcastMessage(observer.UpdateMessage{
			Action: "del", Key: k,
		})
	} else {
		fmt.Fprintf(conn, ":0\r\n")
	}
}

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

func handleClusterMeet(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager) {
	if len(parts) != 4 {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'CLUSTER MEET'\r\n")
		return
	}

	host := parts[2]
	port := parts[3]

	logger.Info("cluster meet requested", "ip", host, "port", port)

	// Add a new node to the cluster (simplified - normally you'd connect and exchange info)
	cm.AddNode(host, port)

	logger.Info("node added to cluster", "total-nodes", len(cm.Nodes))

	fmt.Fprintf(conn, "+OK\r\n")
}

func handleClusterNodes(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager) {
	// TODO: Return cluster topology information
	fmt.Fprintf(conn, "+TODO: cluster nodes not implemented\r\n")
}

func handleClusterInfo(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub, logger *slog.Logger, cm *cluster.Manager) {
	// TODO: Return cluster status information
	fmt.Fprintf(conn, "+TODO: cluster info not implemented\r\n")
}
