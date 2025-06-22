package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
)

func startTestServer(t *testing.T) string {
	// For tests, we can discard log output to keep the test runner clean.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := store.NewStore()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	go func() {
		if err := server.StartWithListener(ln, s, logger); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
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
