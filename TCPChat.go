package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	// Command-line arguments passed to the program
	args := os.Args
	// The port number the server will listen on
	port, err := getPort(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	// The type of network protocol and the full address (e.g., ":8989")
	network := "tcp"
	address := ":" + port

	// The object that waits for new incoming network connections
	listener, err := net.Listen(network, address)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
		return
	}
	fmt.Println("Listening on the port", address)

	// Map to keep track of all currently connected clients
	clients := make(map[net.Conn]bool)
	// Mutex key to prevent data races when multiple goroutines access the clients map
	var key sync.Mutex

	for {
		// A new individual connection to a single client
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		} else {
			key.Lock()
			if len(clients) >= 10 {
				fmt.Fprintln(conn, "Server busy")
				conn.Close()
				key.Unlock()
				continue
			} else {
				clients[conn] = true
				key.Unlock()
				go handleClient(conn, clients, &key)
			}
		}

	}
}

// Handling the client(conn after listener.Accept is valid)
func handleClient(conn net.Conn, clients map[net.Conn]bool, key *sync.Mutex) {
	defer func() {
		key.Lock()
		delete(clients, conn)
		key.Unlock()
		conn.Close()
	}()

	fmt.Fprintln(conn, "[ENTER YOUR NAME]")
	// Scanner to easily read text line-by-line from the client's connection
	msgscanner := bufio.NewScanner(conn)
	msgscanner.Scan()
	// The username chosen by the client
	name := msgscanner.Text()
	for name == "" {
		fmt.Fprintln(conn, "Invalid name, try again")
		msgscanner.Scan()
		name = msgscanner.Text()
	}
	fmt.Fprintln(conn, "Welcome", name)
	key.Lock()
	// Announce the new user to everyone else (receivers) in the chat
	for receiver := range clients {
		if receiver != conn {
			fmt.Fprintln(receiver, "["+time.Now().Format("2006-01-02 15:04:05")+"]", "["+name+"]", "has joined our chat...")
		}
	}
	key.Unlock()

	// Continuously wait for new messages from this client
	for msgscanner.Scan() {
		// The actual chat message typed by the client
		msg := msgscanner.Text()
		key.Lock()
		for receiver := range clients {
			if receiver != conn {
				fmt.Fprintln(receiver, "["+time.Now().Format("2006-01-02 15:04:05")+"]", "["+name+"]", msg)
			}
		}
		key.Unlock()
	}
}

// Reading the correct port
func getPort(args []string) (string, error) {
	if len(args) == 1 {
		return "8989", nil
	}

	if len(args) == 2 {
		return args[1], nil
	}

	return "", errors.New("[USAGE]: ./TCPChat $port")
}
