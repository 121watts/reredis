// Package observer provides real-time WebSocket communication for live data updates.
// This enables modern web applications to receive instant notifications when data changes,
// supporting reactive UIs and collaborative features without polling overhead.
package observer

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// CommandMessage represents a client request sent via WebSocket.
// This enables clients to perform Redis operations through the WebSocket interface,
// providing an alternative to the TCP protocol for web applications.
type CommandMessage struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
}

// UpdateMessage represents a data change notification sent to clients.
// This informs all connected clients about store modifications in real-time,
// enabling reactive user interfaces and live collaboration features.
type UpdateMessage struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
}

// ClusterStatsMessage represents cluster-wide statistics sent to clients.
// This provides real-time monitoring data for cluster health, key distribution,
// and node status updates without requiring polling from the frontend.
type ClusterStatsMessage struct {
	Action        string             `json:"action"`
	Nodes         []ClusterNodeStats `json:"nodes"`
	CurrentNodeID string             `json:"currentNodeId"`
	TotalSlots    int                `json:"totalSlots"`
	ClusterSize   int                `json:"clusterSize"`
	TotalKeys     int                `json:"totalKeys"`
}

// ClusterNodeStats represents statistics for a single cluster node.
// This provides real-time monitoring data for individual nodes including
// key counts, slot assignments, byte usage, and connectivity status.
type ClusterNodeStats struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	SlotStart int32  `json:"slotStart"`
	SlotEnd   int32  `json:"slotEnd"`
	KeyCount  int    `json:"keyCount"`
	ByteSize  int64  `json:"byteSize"`
}

// Hub manages WebSocket connections and broadcasts updates to all clients.
// This centralizes connection management and message distribution, ensuring
// reliable delivery of real-time updates across all connected sessions.
type Hub struct {
	clients    map[*websocket.Conn]bool // Active WebSocket connections for broadcasting
	mu         sync.Mutex               // Protects concurrent access to client map
	broadcast  chan []byte              // Channel for queuing messages to all clients
	logger     *slog.Logger             // Structured logging for debugging and monitoring
	register   chan *websocket.Conn     // Channel for adding new client connections
	unregister chan *websocket.Conn     // Channel for removing disconnected clients
}

// NewHub creates a new WebSocket hub for managing real-time connections.
// This initializes the channels and data structures needed for concurrent
// connection management and message broadcasting in a high-performance manner.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		logger:     logger.With("component", "hub"),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run starts the hub's main event loop for handling client and message events.
// This processes registration, disconnection, and broadcasting in a single goroutine
// to ensure thread safety and prevent race conditions in connection management.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("client registered")
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			h.logger.Info("client unregistered")
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					h.logger.Error("failed to broadcast message", "error", err)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastMessage sends an update notification to all connected WebSocket clients.
// This enables real-time data synchronization by immediately notifying all clients
// about store changes, supporting reactive applications and live collaboration.
func (h *Hub) BroadcastMessage(msg UpdateMessage) {
	bytes, err := json.Marshal(msg)

	if err != nil {
		h.logger.Error("failed to marshal update message", "error", err)
		return
	}

	h.broadcast <- bytes
}

// BroadcastClusterStats sends cluster statistics to all connected WebSocket clients.
// This provides real-time monitoring data without requiring clients to poll for updates.
func (h *Hub) BroadcastClusterStats(stats ClusterStatsMessage) {
	data, err := json.Marshal(stats)
	if err != nil {
		h.logger.Error("failed to marshal cluster stats", "error", err)
		return
	}
	
	h.broadcast <- data
}

// Register adds a new WebSocket connection to receive broadcasts.
// This enables clients to subscribe to real-time updates and participate
// in collaborative features by joining the notification system.
func (h *Hub) Register(client *websocket.Conn) {
	h.register <- client
}

// Unregister removes a WebSocket connection from receiving broadcasts.
// This handles client disconnections gracefully and prevents resource leaks
// by cleaning up inactive connections from the notification system.
func (h *Hub) Unregister(client *websocket.Conn) {
	h.unregister <- client
}
