package main

import (
	"fmt"
	"net"
	"os"
	"sync"

	"net-cat/internal"
)

func main() {
	// Initialize the server and obtain the listener object
	listener, err := internal.StartServer(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Map to keep track of all currently connected clients
	clients := make(map[net.Conn]bool)
	// Mutex key to prevent data races when multiple goroutines access the clients map
	var key sync.Mutex
	broadcast := make(chan string)

	var history []string

	// Start the central broadcaster goroutine
	go broadcaster(broadcast, clients, &key)

	// Main server loop to continuously accept incoming connections
	for {
		// A new individual connection to a single client
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		} else {
			key.Lock()
			// Enforce the maximum connection limit (Task requirement: Max 10 connections)
			if len(clients) >= 10 {
				fmt.Fprintln(conn, "Server is on it's full capacity, try again later")
				conn.Close()
				key.Unlock()
				continue
			} else {
				// Add the new client to the active clients map
				clients[conn] = true
				key.Unlock()

				// Start a new goroutine to handle this client concurrently so the loop isn't blocked
				go internal.HandleClient(conn, clients, &key, broadcast, &history)
			}
		}

	}
}

func broadcaster(broadcast chan string, clients map[net.Conn]bool, key *sync.Mutex) {
	for {
		msg := <-broadcast
		key.Lock()
		for receiver := range clients {
			fmt.Fprintln(receiver, msg)
		}
		key.Unlock()
	}
}
