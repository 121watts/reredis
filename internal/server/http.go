package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/121watts/reredis/internal/observer"
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

func handleWsConnection(hub *observer.Hub, s *store.Store, w http.ResponseWriter, r *http.Request) {
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
		case "DEL":
			if s.Delete(cmd.Key) {
				hub.BroadcastMessage(observer.UpdateMessage{
					Action: "del", Key: cmd.Key,
				})
			}
		}
	}
}

// NewHTTPHandler creates the main http handler for the web server.
func NewHTTPHandler(hub *observer.Hub, s *store.Store, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWsConnection(hub, s, w, r)
	})
	return mux
}

func StartWebServer(addr string, hub *observer.Hub, logger *slog.Logger, s *store.Store) error {
	handler := NewHTTPHandler(hub, s, logger)
	logger.Info("starting web server for websockets", "addr", addr)
	return http.ListenAndServe(addr, handler)
}
