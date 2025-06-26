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
	"fmt"
	"io"
	"time"
)

// upgrader configures WebSocket connection upgrades with permissive CORS for development.
// This enables real-time communication between web browsers and the server while allowing
// cross-origin requests for modern web application architectures.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for development purposes.
		// In production, you'd want to restrict this.
		return true
	},
}

// ClusterNodeInfo represents node metadata for the WebSocket cluster dashboard.
// This provides visibility into cluster topology and data distribution,
// enabling monitoring and debugging of distributed operations.
type ClusterNodeInfo struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	SlotStart int32  `json:"slotStart"`
	SlotEnd   int32  `json:"slotEnd"`
	KeyCount  int    `json:"keyCount"`
	ByteSize  int64  `json:"byteSize"`
}

// ClusterInfoResponse provides comprehensive cluster state information via WebSocket.
// This supports real-time cluster monitoring and administration tools
// with current topology and health metrics.
type ClusterInfoResponse struct {
	Action        string            `json:"action"`
	Nodes         []ClusterNodeInfo `json:"nodes"`
	CurrentNodeID string            `json:"currentNodeId"`
	TotalSlots    int32             `json:"totalSlots"`
	ClusterSize   int               `json:"clusterSize"`
}

// ClusterEventResponse notifies clients about cluster topology changes.
// This enables reactive dashboards and monitoring tools to update their
// views when nodes join, leave, or change state.
type ClusterEventResponse struct {
	Action  string `json:"action"`
	Event   string `json:"event"`
	NodeID  string `json:"nodeId,omitempty"`
	Message string `json:"message"`
}

// handleWsConnection manages individual WebSocket connections for real-time operations.
// This enables web clients to perform Redis commands and receive live updates,
// bridging the gap between traditional Redis clients and modern web applications.
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
				var keyCount int
				var byteSize int64
				
				// For the current node, get the actual stats from the store
				if node.ID == cm.Node.ID {
					keyCount = len(s.GetAll())
					byteSize = s.GetTotalByteSize()
					// Update the cluster manager with the actual counts
					cm.UpdateKeyCount(keyCount)
					cm.UpdateByteSize(byteSize)
				} else {
					// For other nodes, fetch the stats via HTTP
					keyCount = getKeyCountFromNode(node.Host, node.Port)
					byteSize = getByteSizeFromNode(node.Host, node.Port)
					// Update the cluster manager with the fetched counts
					node.KeyCount = keyCount
					node.ByteSize = byteSize
				}

				nodes = append(nodes, ClusterNodeInfo{
					ID:        node.ID,
					Host:      node.Host,
					Port:      node.Port,
					SlotStart: node.Slot.Start,
					SlotEnd:   node.Slot.End,
					KeyCount:  keyCount,
					ByteSize:  byteSize,
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

// handleGetKeys provides paginated key listing via HTTP REST API.
// This supports administrative tools and debugging by enabling efficient
// iteration over large key sets without overwhelming client or server memory.
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

// getKeyCountFromNode fetches the key count from a remote cluster node via HTTP.
// This enables the cluster dashboard to display accurate statistics from all nodes.
func getKeyCountFromNode(host, port string) int {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://%s:%s/keycount", host, getHTTPPort(port))
	
	resp, err := client.Get(url)
	if err != nil {
		return 0 // Return 0 if node is unreachable
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}
	
	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0
	}
	
	return count
}

// getByteSizeFromNode fetches the byte size from a remote cluster node via HTTP.
// This enables the cluster dashboard to display accurate storage statistics from all nodes.
func getByteSizeFromNode(host, port string) int64 {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://%s:%s/bytesize", host, getHTTPPort(port))
	
	resp, err := client.Get(url)
	if err != nil {
		return 0 // Return 0 if node is unreachable
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}
	
	size, err := strconv.ParseInt(strings.TrimSpace(string(body)), 10, 64)
	if err != nil {
		return 0
	}
	
	return size
}

// getHTTPPort converts a TCP port to the corresponding HTTP port.
// This assumes HTTP ports are TCP port + 2700 (e.g., 6379 -> 9079, 6380 -> 9080).
func getHTTPPort(tcpPort string) string {
	port, err := strconv.Atoi(tcpPort)
	if err != nil {
		return "9080" // Default fallback
	}
	return strconv.Itoa(port + 2700)
}

// NewHTTPHandler creates the main HTTP handler with WebSocket and REST endpoints.
// This provides a unified interface for both real-time WebSocket operations
// and traditional HTTP APIs, supporting diverse client needs and integration patterns.
func NewHTTPHandler(hub *observer.Hub, s *store.Store, cm *cluster.Manager, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWsConnection(hub, s, cm, w, r)
	})

	mux.HandleFunc("GET /api/v1/keys", func(w http.ResponseWriter, r *http.Request) {
		handleGetKeys(s, w, r)
	})

	// Add keycount endpoint for cluster statistics
	mux.HandleFunc("GET /keycount", func(w http.ResponseWriter, r *http.Request) {
		count := len(s.GetAll())
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%d", count)
	})

	// Add bytesize endpoint for cluster storage statistics
	mux.HandleFunc("GET /bytesize", func(w http.ResponseWriter, r *http.Request) {
		size := s.GetTotalByteSize()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%d", size)
	})

	return mux
}

// StartWebServer launches the HTTP server for WebSocket and REST API access.
// This enables web-based clients and dashboards to interact with the Redis-compatible
// store through modern protocols while maintaining compatibility with existing tools.
func StartWebServer(addr string, hub *observer.Hub, logger *slog.Logger, s *store.Store, cm *cluster.Manager) error {
	handler := NewHTTPHandler(hub, s, cm, logger)
	logger.Info("starting web server for websockets", "addr", addr)
	return http.ListenAndServe(addr, handler)
}
