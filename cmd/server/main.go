package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	defer listener.Close()
	log.Println("Listening on :6379")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("Failed to accept connection %v", err)
			continue
		}

		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Received: %s\n", line)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
}
