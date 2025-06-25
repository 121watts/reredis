package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/121watts/reredis/internal/cluster"
	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/query"
	"github.com/121watts/reredis/internal/store"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for development purposes.
		// In production, you'd want to restrict this.
		return true
	},
}

type ClusterNodeInfo struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	SlotStart int32  `json:"slotStart"`
	SlotEnd   int32  `json:"slotEnd"`
	KeyCount  int    `json:"keyCount"`
}

type ClusterInfoResponse struct {
	Action        string            `json:"action"`
	Nodes         []ClusterNodeInfo `json:"nodes"`
	CurrentNodeID string            `json:"currentNodeId"`
	TotalSlots    int32             `json:"totalSlots"`
	ClusterSize   int               `json:"clusterSize"`
}

type ClusterEventResponse struct {
	Action  string `json:"action"`
	Event   string `json:"event"`
	NodeID  string `json:"nodeId,omitempty"`
	Message string `json:"message"`
}

func handleWsConnection(hub *observer.Hub, s *store.Store, cm *cluster.Manager, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		slog.Error("failed to upgrade connection", "error", err)
		return
	}

	defer ws.Close()
	hub.Register(ws)
	defer hub.Unregister(ws)

	for {
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("error reading from websocket", "error", err)
			}
			break
		}

		var cmd observer.CommandMessage

		if err := json.Unmarshal(msgBytes, &cmd); err != nil {
			slog.Error("failed to unmarshal command", "error", err)
			continue
		}

		switch strings.ToUpper(cmd.Action) {
		case "SET":
			s.Set(cmd.Key, cmd.Value)
			hub.BroadcastMessage(observer.UpdateMessage{
				Action: "set", Key: cmd.Key, Value: cmd.Value,
			})
		case "GET":
			val, ok := s.Get(cmd.Key)
			if !ok {
				val = "(nil)" // Or some other indicator of not found
			}
			resp := observer.UpdateMessage{Action: "get_resp", Key: cmd.Key, Value: val}
			if err := ws.WriteJSON(resp); err != nil {
				slog.Error("failed to send GET response", "error", err)
			}
		case "GET_ALL":
			allData := s.GetAll()
			resp := struct {
				Action string            `json:"action"`
				Data   map[string]string `json:"data"`
			}{Action: "sync", Data: allData}

			if err := ws.WriteJSON(resp); err != nil {
				slog.Error("failed to send sync response", "error", err)
			}
		case "DEL":
			if s.Delete(cmd.Key) {
				hub.BroadcastMessage(observer.UpdateMessage{
					Action: "del", Key: cmd.Key,
				})
			}
		case "CLUSTER_INFO":
			// Create cluster info response
			nodes := make([]ClusterNodeInfo, 0, len(cm.Nodes))
			for _, node := range cm.Nodes {
				keyCount := 0
				// Count keys belonging to this node (simplified - count all keys for current node)
				if node.ID == cm.Node.ID {
					keyCount = len(s.GetAll())
				}

				nodes = append(nodes, ClusterNodeInfo{
					ID:        node.ID,
					Host:      node.Host,
					Port:      node.Port,
					SlotStart: node.Slot.Start,
					SlotEnd:   node.Slot.End,
					KeyCount:  keyCount,
				})
			}

			resp := ClusterInfoResponse{
				Action:        "cluster_info",
				Nodes:         nodes,
				CurrentNodeID: cm.Node.ID,
				TotalSlots:    cluster.SLOT_RANGE,
				ClusterSize:   len(cm.Nodes),
			}

			if err := ws.WriteJSON(resp); err != nil {
				slog.Error("failed to send cluster info response", "error", err)
			}
		}
	}
}

func handleGetKeys(s *store.Store, w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	rLimit := r.URL.Query().Get("limit")
	cursor := r.URL.Query().Get("cursor")

	limit, err := strconv.Atoi(rLimit)
	if err != nil {
		limit = 20
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	resp := query.HandleCursorPagination(s, cursor, limit)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// NewHTTPHandler creates the main http handler for the web server.
func NewHTTPHandler(hub *observer.Hub, s *store.Store, cm *cluster.Manager, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWsConnection(hub, s, cm, w, r)
	})

	mux.HandleFunc("GET /api/v1/keys", func(w http.ResponseWriter, r *http.Request) {
		handleGetKeys(s, w, r)
	})

	return mux
}

func StartWebServer(addr string, hub *observer.Hub, logger *slog.Logger, s *store.Store, cm *cluster.Manager) error {
	handler := NewHTTPHandler(hub, s, cm, logger)
	logger.Info("starting web server for websockets", "addr", addr)
	return http.ListenAndServe(addr, handler)
}
