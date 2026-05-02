package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"net-cat/internal"
)

func main() {
	// Open a file to store all the server and chat logs
	logFile, err := os.OpenFile("server_logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
	}
	defer logFile.Close()

	// Use io.MultiWriter to write logs to BOTH the terminal and the file
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

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
