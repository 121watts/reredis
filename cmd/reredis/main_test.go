package main

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
)

func startTestServer(t *testing.T) string {
	s := store.NewStore()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	go func() {
		if err := server.StartWithListener(ln, s); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return addr
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

func TestIntegrationSetAndGet(t *testing.T) {
	addr := startTestServer(t)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("could not connect to server: %v", err)
	}
	defer conn.Close()

	setResp := sendCommand(t, conn, "SET foo bar")
	if setResp != "+OK\r\n" {
		t.Errorf("expected +OK, got %q", setResp)
	}

	getResp := sendCommand(t, conn, "GET foo")
	if getResp != "bar\r\n" {
		t.Errorf("expected 'bar\\r\\n', got %q", getResp)
	}

	getNilResp := sendCommand(t, conn, "GET baz")
	if getNilResp != "-ERR key not found\r\n" {
		t.Errorf("expected '-ERR key not found\\r\\n', got %q", getNilResp)
	}
}
