// Package main provides the entry point for the Reredis server.
// This lightweight Redis-compatible server supports clustering, real-time WebSocket
// updates, and TTL-based key expiration for distributed applications.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/121watts/reredis/internal/cluster"
	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
)

// main initializes and starts the Reredis server with both TCP and HTTP interfaces.
// This enables Redis protocol compatibility for existing applications while providing
// WebSocket support for real-time features and modern web applications.
func main() {
	tcpPort := flag.Int("port", 6379, "TCP port for Redis protocol")
	httpPort := flag.Int("http-port", 8080, "HTTP port for WebSocket connections")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create cluster manager
	cm := cluster.NewManager("127.0.0.1", fmt.Sprintf("%d", *tcpPort))

	tcpAddr := fmt.Sprintf(":%d", *tcpPort)
	httpAddr := fmt.Sprintf(":%d", *httpPort)

	logger.Info("starting cluster node", "node-id", cm.Node.ID, "tcp-port", *tcpPort, "http-port", *httpPort, "slot-range", cm.Node.Slot)

	s := store.NewStore()
	hub := observer.NewHub(logger)
	go hub.Run()

	go func() {
		if err := server.Start(tcpAddr, s, logger, hub, cm); err != nil {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	if err := server.StartWebServer(httpAddr, hub, logger, s, cm); err != nil {
		logger.Error("http server failed", "error", err)
		os.Exit(1)
	}
}
