package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
	"github.com/gorilla/websocket"
)

func startTestServer(t *testing.T) string {
	// For tests, we can discard log output to keep the test runner clean.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := observer.NewHub(logger)
	go hub.Run()
	s := store.NewStore()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	go func() {
		if err := server.StartWithListener(ln, s, logger, hub); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	const maxRetries = 10
	const retryDelay = 100 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return addr
		}
		time.Sleep(retryDelay)
	}
	t.Fatalf("server did not start in time")
	return addr
}

func newConn(t *testing.T, addr string) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("could not connect to server: %v", err)
	}
	return conn
}

func sendCommand(t *testing.T, conn net.Conn, cmd string) string {
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

func newWsConn(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	// httptest.Server uses http, but websocket needs ws.
	wsURL := "ws" + strings.TrimPrefix(url, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not connect to websocket: %v", err)
	}
	t.Cleanup(func() {
		conn.Close()
	})
	return conn
}

func TestIntegration(t *testing.T) {
	addr := startTestServer(t)

	t.Run("SET and GET", func(t *testing.T) {
		conn := newConn(t, addr)
		defer conn.Close()

		setResp := sendCommand(t, conn, "SET foo bar")
		if setResp != "+OK\r\n" {
			t.Errorf("expected +OK, got %q", setResp)
		}

		getResp := sendCommand(t, conn, "GET foo")
		if getResp != "bar\r\n" {
			t.Errorf("expected 'bar\\r\\n', got %q", getResp)
		}
	})

	t.Run("GET non-existent key", func(t *testing.T) {
		conn := newConn(t, addr)
		defer conn.Close()
		getNilResp := sendCommand(t, conn, "GET baz")
		if getNilResp != "-ERR key not found\r\n" {
			t.Errorf("expected '-ERR key not found\\r\\n', got %q", getNilResp)
		}
	})

	t.Run("DEL command", func(t *testing.T) {
		conn := newConn(t, addr)
		defer conn.Close()

		// Set a key to ensure it exists before deleting
		sendCommand(t, conn, "SET key-to-del value")

		// Test deleting an existing key
		delResp := sendCommand(t, conn, "DEL key-to-del")
		if delResp != ":1\r\n" {
			t.Errorf("expected ':1\\r\\n' when deleting existing key, got %q", delResp)
		}

		// Verify the key is gone
		getAfterDelResp := sendCommand(t, conn, "GET key-to-del")
		if getAfterDelResp != "-ERR key not found\r\n" {
			t.Errorf("expected key not found after delete, got %q", getAfterDelResp)
		}

		// Test deleting a non-existent key
		delNonExistentResp := sendCommand(t, conn, "DEL non-existent-key")
		if delNonExistentResp != ":0\r\n" {
			t.Errorf("expected ':0\\r\\n' when deleting non-existent key, got %q", delNonExistentResp)
		}
	})
}

func TestWebsocketIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := observer.NewHub(logger)
	go hub.Run()
	s := store.NewStore()

	// Start an httptest server for WebSockets
	httpHandler := server.NewHTTPHandler(hub, s, logger)
	httpServer := httptest.NewServer(httpHandler)
	t.Cleanup(httpServer.Close)

	t.Run("SET command via WebSocket broadcasts correctly", func(t *testing.T) {
		// Create two clients to check broadcast logic
		clientA := newWsConn(t, httpServer.URL)
		clientB := newWsConn(t, httpServer.URL)

		// Client A sends a SET command
		cmd := observer.CommandMessage{Action: "set", Key: "ws-test", Value: "success"}
		if err := clientA.WriteJSON(cmd); err != nil {
			t.Fatalf("failed to send command: %v", err)
		}

		// Expect Client B to receive a broadcast
		var receivedMsg observer.UpdateMessage
		if err := clientB.ReadJSON(&receivedMsg); err != nil {
			t.Fatalf("failed to read broadcast message: %v", err)
		}

		if receivedMsg.Action != "set" || receivedMsg.Key != "ws-test" || receivedMsg.Value != "success" {
			t.Errorf("incorrect broadcast message received: got %+v", receivedMsg)
		}
	})

	t.Run("GET command via WebSocket returns value", func(t *testing.T) {
		client := newWsConn(t, httpServer.URL)
		s.Set("get-key", "get-value") // Pre-populate the store

		cmd := observer.CommandMessage{Action: "get", Key: "get-key"}
		if err := client.WriteJSON(cmd); err != nil {
			t.Fatalf("failed to send GET command: %v", err)
		}

		var receivedMsg observer.UpdateMessage
		if err := client.ReadJSON(&receivedMsg); err != nil {
			t.Fatalf("failed to read GET response: %v", err)
		}

		if receivedMsg.Action != "get_resp" || receivedMsg.Key != "get-key" || receivedMsg.Value != "get-value" {
			t.Errorf("incorrect GET response received: got %+v", receivedMsg)
		}
	})
}
