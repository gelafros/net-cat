package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"net-cat/internal"
)

func main() {
	// Setup logging to both the terminal and the log file
	logFile := internal.SetupLogging()
	defer logFile.Close()

	// Initialize the server and obtain the listener object
	listener, err := internal.StartServer(os.Args)
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}

	// Map to keep track of all currently connected clients
	clients := make(map[net.Conn]internal.Client)
	// Mutex key to prevent data races when multiple goroutines access the clients map
	var key sync.Mutex

	// Channel for central message broadcasting
	broadcast := make(chan internal.Message)

	// Slice to hold the chat history so new clients can instantly catch up
	var history []string

	// Start the central broadcaster goroutine
	go internal.Broadcaster(broadcast, clients, &key)

	// Main server loop to continuously accept incoming connections
	for {
		// A new individual connection to a single client
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		} else {
			key.Lock()
			// Enforce the maximum connection limit (Task requirement: Max 10 connections)
			if len(clients) >= 10 {
				if _, err := fmt.Fprintln(conn, "Server is on it's full capacity, try again later"); err != nil {
					log.Printf("Error writing capacity message to %s: %v\n", conn.RemoteAddr().String(), err)
				}
				conn.Close()
				key.Unlock()
				continue
			} else {
				// Add the new client to the active clients map
				clients[conn] = internal.Client{}
				key.Unlock()

				// Start a new goroutine to handle this client concurrently so the loop isn't blocked
				go internal.HandleClient(conn, clients, &key, broadcast, &history)
			}
		}

	}
}
