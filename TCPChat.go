package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"net-cat/internal"
)

func main() {
	// Initialize the server and obtain the listener object
	listener, err := internal.StartServer(os.Args)
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}

	// Map to keep track of all currently connected clients
	clients := make(map[net.Conn]string)
	// Mutex key to prevent data races when multiple goroutines access the clients map
	var key sync.Mutex
	broadcast := make(chan internal.Message)

	// Slice to hold the chat history so new clients can instantly catch up
	var history []string

	// Start the central broadcaster goroutine
	go broadcaster(broadcast, clients, &key)

	// Main server loop to continuously accept incoming connections
	for {
		// A new individual connection to a single client
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		} else {
			log.Printf("New connection accepted from %s\n", conn.RemoteAddr().String())
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
				clients[conn] = ""
				key.Unlock()

				// Start a new goroutine to handle this client concurrently so the loop isn't blocked
				go internal.HandleClient(conn, clients, &key, broadcast, &history)
			}
		}

	}
}

// broadcaster is a background goroutine that continuously listens for incoming
// messages on the broadcast channel and distributes them to all active clients.
func broadcaster(broadcast chan internal.Message, clients map[net.Conn]string, key *sync.Mutex) {
	for {
		// Pause and wait for a new message to be sent into the channel
		msg := <-broadcast

		key.Lock()
		// Send the message to every connected client in the chat room
		for receiver, receiverName := range clients {
			// Only send to clients who have joined (have a name) and aren't the sender
			if receiver != msg.Sender && receiverName != "" {
				if _, err := fmt.Fprintln(receiver, "\r\033[2K"+msg.Text); err != nil {
					log.Printf("Error sending message to %s: %v\n", receiver.RemoteAddr().String(), err)
				}
				fmt.Fprintf(receiver, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), receiverName)
			}
		}
		key.Unlock()
	}
}
