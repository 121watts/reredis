package server

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/store"
)

func Start(address string, s *store.Store, logger *slog.Logger, hub *observer.Hub) error {
	ln, err := net.Listen("tcp", address)

	if err != nil {
		return fmt.Errorf("failed to bind: %w", err)
	}

	return StartWithListener(ln, s, logger, hub)
}

func StartWithListener(ln net.Listener, s *store.Store, logger *slog.Logger, hub *observer.Hub) error {
	defer ln.Close()
	logger.Info("listening on port", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()

		if err != nil {
			logger.Error("failed to accept connection", "error", err)
			continue
		}

		lw := logger.With("remote_addr", conn.RemoteAddr().String())

		go handleConnection(conn, s, lw, hub)
	}
}

func handleConnection(conn net.Conn, s *store.Store, logger *slog.Logger, hub *observer.Hub) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) == 0 {
			fmt.Fprintf(conn, "-ERR empty command\r\n")
			continue
		}

		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "SET":
			handleSet(parts, s, conn, hub)
		case "GET":
			handleGet(parts, s, conn, hub)
		case "DEL":
			handleDelete(parts, s, conn, hub)
		default:
			fmt.Fprintf(conn, "-ERR unknown command\r\n")
		}

	}

	if err := scanner.Err(); err != nil {
		logger.Error("error reading from connection", "error", err)
	}
}

func handleSet(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub) {
	const expectedParts = 3
	if len(parts) != expectedParts {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'SET'\r\n")
		return
	}

	k, v := parts[1], parts[2]
	s.Set(k, v)
	hub.BroadcastMessage(observer.UpdateMessage{
		Action: "set", Key: k, Value: v,
	})
	fmt.Fprintf(conn, "+OK\r\n")
}

func handleGet(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub) {
	const expectedParts = 2

	if len(parts) != expectedParts {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'GET'\r\n")
		return
	}

	k := parts[1]

	v, ok := s.Get(k)

	if !ok {
		fmt.Fprintf(conn, "-ERR key not found\r\n")
		return
	}

	hub.BroadcastMessage(observer.UpdateMessage{
		Action: "get", Key: k,
	})

	fmt.Fprintf(conn, "%s\r\n", v)
}

func handleDelete(parts []string, s *store.Store, conn net.Conn, hub *observer.Hub) {
	const expectedParts = 2

	if len(parts) != expectedParts {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'DEL'\r\n")
		return
	}

	k := parts[1]

	ok := s.Delete(k)

	if ok {
		fmt.Fprintf(conn, ":1\r\n")
		hub.BroadcastMessage(observer.UpdateMessage{
			Action: "del", Key: k,
		})
	} else {
		fmt.Fprintf(conn, ":0\r\n")
	}
}
