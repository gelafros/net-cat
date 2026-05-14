package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"net-cat/cmd/client"
	"net-cat/internal"
)

func main() {
	isUI := false
	isClient := false
	host := "localhost"
	port := "8989"
	args := os.Args[1:]

	// If the first argument looks like a port, and there are more arguments,
	// it's an invalid server command format.
	if len(args) > 1 && isPort(args[0]) {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	// Parse arguments to determine the execution mode
	for _, arg := range args {
		if arg == "-ui" {
			isUI = true
			isClient = true
		} else if arg == "localhost" || strings.Contains(arg, ".") {
			host = arg
			isClient = true
		} else {
			port = arg
		}
	}

	// Mode 1: UI Client
	if isUI && isClient {
		client.Run(host + ":" + port)
		return
	}

	// Mode 2: Standard CLI Client (Mimicking system 'nc')
	if !isUI && isClient {
		runCLIClient(host + ":" + port)
		return
	}

	// Setup logging to both the terminal and the log file
	logFile := internal.SetupLogging()
	defer logFile.Close()

	// Initialize the server and obtain the listener object
	listener, err := internal.StartServer(os.Args)
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}

	manager := internal.NewRoomManager()

	// Main server loop to continuously accept incoming connections
	for {
		// A new individual connection to a single client
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		} else {
			manager.Mutex.Lock()
			// Enforce the maximum connection limit (Task requirement: Max 10 connections)
			if manager.TotalClients >= 10 {
				manager.Mutex.Unlock()
				if _, err := fmt.Fprintln(conn, "Server is on its full capacity, try again later"); err != nil {
					log.Printf("Error writing capacity message to %s: %v\n", conn.RemoteAddr().String(), err)
				}
				conn.Close()
				continue
			} else {
				manager.TotalClients++
				manager.Mutex.Unlock()

				// Start a new goroutine to handle this client concurrently so the loop isn't blocked
				go internal.HandleClient(conn, manager)
			}
		}

	}
}

func runCLIClient(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()
	go io.Copy(os.Stdout, conn)
	io.Copy(conn, os.Stdin)
}

// Simple helper to check if a string consists only of digits
func isPort(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
