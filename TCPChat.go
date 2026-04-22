package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	"net-cat/handlers"
)

// Handling Arguments (correct port)
func getPort(args []string) (string, error) {
	if len(args) == 1 {
		return "8989", nil
	}

	if len(args) == 2 {
		return args[1], nil
	}

	return "", errors.New("[USAGE]: ./TCPChat $port")
}

func main() {
	args := os.Args
	port, err := getPort(args)
	if err != nil {
		fmt.Println("Error in getting port.", err)
		return
	}
	network := "tcp"
	address := ":" + port

	// Create a server in port "address". Listener is a handler for the server
	listener, err := net.Listen(network, address)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
		return
	}
	fmt.Println("Listening on the port", address)
	server := handlers.NewServer()

	// Running loop for receiving multiple clients (Accept loop)
	for {
		client, err := listener.Accept()
		if err != nil {
			fmt.Println("Error with client connection:", err)
			continue // for not terminating the Server
		}
		go handlers.HandleClient(client, server)
	}
}
