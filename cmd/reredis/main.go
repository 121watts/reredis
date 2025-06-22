package main

import (
	"log/slog"
	"os"

	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	s := store.NewStore()
	addr := ":6379"
	logger.Info("starting reredis server", "addr", addr)

	if err := server.Start(addr, s, logger); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
