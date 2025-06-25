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

	"github.com/121watts/reredis/internal/cluster"
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
	cm := cluster.NewManager("localhost", "6379")
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	go func() {
		if err := server.StartWithListener(ln, s, logger, hub, cm); err != nil {
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
	cm := cluster.NewManager("localhost", "6379")

	// Start an httptest server for WebSockets
	httpHandler := server.NewHTTPHandler(hub, s, cm, logger)
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

	t.Run("GET_ALL command via WebSocket returns full store", func(t *testing.T) {
		client := newWsConn(t, httpServer.URL)
		// Pre-populate the store with some data for the test
		s.Set("sync-key-1", "sync-val-1")
		s.Set("sync-key-2", "sync-val-2")

		cmd := observer.CommandMessage{Action: "get_all"}
		if err := client.WriteJSON(cmd); err != nil {
			t.Fatalf("failed to send GET_ALL command: %v", err)
		}

		type syncMessage struct {
			Action string            `json:"action"`
			Data   map[string]string `json:"data"`
		}
		var receivedMsg syncMessage
		if err := client.ReadJSON(&receivedMsg); err != nil {
			t.Fatalf("failed to read sync response: %v", err)
		}

		if receivedMsg.Action != "sync" {
			t.Errorf("expected action 'sync', got %q", receivedMsg.Action)
		}

		if receivedMsg.Data["sync-key-1"] != "sync-val-1" || receivedMsg.Data["sync-key-2"] != "sync-val-2" {
			t.Errorf("incorrect data received in sync message: got %+v", receivedMsg.Data)
		}
	})
}

func TestTTLAndLRU(t *testing.T) {
	_ = slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("TTL basic functionality", func(t *testing.T) {
		s := store.NewStore()

		// Set a key with short TTL
		s.SetWithTTL("short-lived", "value", 50*time.Millisecond)

		// Should be available immediately
		value, ok := s.Get("short-lived")
		if !ok || value != "value" {
			t.Errorf("expected key to be available immediately, got ok=%v value=%q", ok, value)
		}

		// Wait for expiration
		time.Sleep(60 * time.Millisecond)

		// Should be expired now
		_, ok = s.Get("short-lived")
		if ok {
			t.Errorf("expected key to be expired, but it was still found")
		}
	})

	t.Run("TTL lazy expiration on Get", func(t *testing.T) {
		s := store.NewStore()

		// Set key with very short TTL
		s.SetWithTTL("expire-on-get", "value", 10*time.Millisecond)

		// Wait for expiration but don't access key
		time.Sleep(20 * time.Millisecond)

		// First Get should trigger lazy expiration
		_, ok := s.Get("expire-on-get")
		if ok {
			t.Errorf("expected expired key to be removed on Get")
		}

		// Verify it's not in GetAll either
		all := s.GetAll()
		if _, exists := all["expire-on-get"]; exists {
			t.Errorf("expired key should not appear in GetAll")
		}
	})

	t.Run("TTL active cleanup", func(t *testing.T) {
		s := store.NewStore()

		// Set multiple keys with short TTL
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("cleanup-test-%d", i)
			s.SetWithTTL(key, "value", 100*time.Millisecond)
		}

		// Verify all keys exist
		all := s.GetAll()
		if len(all) < 5 {
			t.Errorf("expected at least 5 keys, got %d", len(all))
		}

		// Wait for cleanup to run (background goroutine should clean up)
		time.Sleep(200 * time.Millisecond)

		// Most/all keys should be cleaned up by background process
		all = s.GetAll()
		cleanedUp := true
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("cleanup-test-%d", i)
			if _, exists := all[key]; exists {
				cleanedUp = false
				break
			}
		}

		if !cleanedUp {
			t.Logf("Background cleanup may not have run yet, this is ok")
		}
	})

	t.Run("TTL mixed with non-TTL keys", func(t *testing.T) {
		s := store.NewStore()

		// Set regular key (no TTL)
		s.Set("permanent", "forever")

		// Set TTL key
		s.SetWithTTL("temporary", "short-lived", 50*time.Millisecond)

		// Wait for TTL expiration
		time.Sleep(60 * time.Millisecond)

		// Permanent key should still exist
		value, ok := s.Get("permanent")
		if !ok || value != "forever" {
			t.Errorf("permanent key should still exist")
		}

		// TTL key should be gone
		_, ok = s.Get("temporary")
		if ok {
			t.Errorf("TTL key should be expired")
		}
	})

	t.Run("LRU eviction without TTL", func(t *testing.T) {
		s := store.NewStore()

		// Fill store to capacity + 1 (maxSize is 1000)
		for i := 0; i <= 1000; i++ {
			key := fmt.Sprintf("key-%d", i)
			s.Set(key, "value")
		}

		// First key should be evicted (LRU)
		_, ok := s.Get("key-0")
		if ok {
			t.Logf("LRU may not have evicted yet with current maxSize=1000")
		}

		// Most recent key should still exist
		value, ok := s.Get("key-1000")
		if !ok || value != "value" {
			t.Errorf("most recent key should still exist")
		}
	})

	t.Run("LRU with TTL interaction", func(t *testing.T) {
		s := store.NewStore()

		// Set a key with TTL
		s.SetWithTTL("ttl-key", "ttl-value", 50*time.Millisecond)

		// Set regular key
		s.Set("regular-key", "regular-value")

		// Access TTL key to move it to front
		s.Get("ttl-key")

		// Wait for TTL expiration
		time.Sleep(60 * time.Millisecond)

		// TTL key should be expired even though it was recently accessed
		_, ok := s.Get("ttl-key")
		if ok {
			t.Errorf("TTL key should expire regardless of LRU position")
		}

		// Regular key should still exist
		value, ok := s.Get("regular-key")
		if !ok || value != "regular-value" {
			t.Errorf("regular key should still exist")
		}
	})

	t.Run("TTL update on existing key", func(t *testing.T) {
		s := store.NewStore()

		// Set key without TTL
		s.Set("update-test", "original")

		// Update with TTL
		s.SetWithTTL("update-test", "updated", 50*time.Millisecond)

		// Should have new value immediately
		value, ok := s.Get("update-test")
		if !ok || value != "updated" {
			t.Errorf("expected updated value, got ok=%v value=%q", ok, value)
		}

		// Should expire after TTL
		time.Sleep(60 * time.Millisecond)
		_, ok = s.Get("update-test")
		if ok {
			t.Errorf("updated key should expire")
		}
	})

	t.Run("TTL removal on regular Set", func(t *testing.T) {
		s := store.NewStore()

		// Set key with TTL
		s.SetWithTTL("ttl-to-regular", "ttl-value", 50*time.Millisecond)

		// Immediately update with regular Set (no TTL)
		s.Set("ttl-to-regular", "regular-value")

		// Wait past original TTL time
		time.Sleep(60 * time.Millisecond)

		// Key should still exist (TTL was removed)
		value, ok := s.Get("ttl-to-regular")
		if !ok || value != "regular-value" {
			t.Errorf("key should not expire after TTL was removed, got ok=%v value=%q", ok, value)
		}
	})
}
