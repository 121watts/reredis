package main

import (
	"log"

	"github.com/121watts/reredis/internal/server"
	"github.com/121watts/reredis/internal/store"
)

func main() {
	s := store.NewStore()
	if err := server.Start(":6379", s); err != nil {
		log.Fatal(err)
	}
}
