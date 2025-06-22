package server

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/121watts/reredis/internal/store"
)

func Start(address string, s *store.Store, logger *slog.Logger) error {
	ln, err := net.Listen("tcp", address)

	if err != nil {
		return fmt.Errorf("failed to bind: %w", err)
	}

	return StartWithListener(ln, s, logger)
}

func StartWithListener(ln net.Listener, s *store.Store, logger *slog.Logger) error {
	defer ln.Close()
	logger.Info("listening on port", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()

		if err != nil {
			logger.Error("failed to accept connection", "error", err)
			continue
		}

		go handleConnection(conn, s, logger.With("remote_addr", conn.RemoteAddr().String()))
	}
}

func handleConnection(conn net.Conn, s *store.Store, logger *slog.Logger) {
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
			handleSet(parts, s, conn)
		case "GET":
			handleGet(parts, s, conn)
		default:
			fmt.Fprintf(conn, "-ERR unknown command\r\n")
		}

	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
}

func handleSet(parts []string, s *store.Store, conn net.Conn) {
	const expectedParts = 3
	if len(parts) != expectedParts {
		fmt.Fprintf(conn, "-ERR wrong number of arguments for 'SET'\r\n")
		return
	}

	k, v := parts[1], parts[2]
	s.Set(k, v)
	fmt.Fprintf(conn, "+OK\r\n")
}

func handleGet(parts []string, s *store.Store, conn net.Conn) {
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

	fmt.Fprintf(conn, "%s\r\n", v)
}
