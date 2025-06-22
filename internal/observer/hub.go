package observer

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type CommandMessage struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
}

type UpdateMessage struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	mu         sync.Mutex
	broadcast  chan []byte
	logger     *slog.Logger
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		logger:     logger.With("component", "hub"),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

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

func (h *Hub) BroadcastMessage(msg UpdateMessage) {
	bytes, err := json.Marshal(msg)

	if err != nil {
		h.logger.Error("failed to marshal update message", "error", err)
		return
	}

	h.broadcast <- bytes
}

func (h *Hub) Register(client *websocket.Conn) {
	h.register <- client
}

func (h *Hub) Unregister(client *websocket.Conn) {
	h.unregister <- client
}
