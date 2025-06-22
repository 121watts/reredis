package main

import (
	"log/slog"
	"os"

	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	s := store.NewStore()
	hub := observer.NewHub(logger)
	go hub.Run()
	tcpAddr := ":6379"
	httpAddr := ":8080"

	go func() {
		if err := server.Start(tcpAddr, s, logger, hub); err != nil {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	if err := server.StartWebServer(httpAddr, hub, logger, s); err != nil {
		logger.Error("http server failed", "error", err)
		os.Exit(1)
	}
}
