package handlers

import (
	"errors"
	"fmt"
	"net"
)

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

func StartServer(args []string) (net.Listener, error) {
	// The port number the server will listen on
	port, err := getPort(args)
	if err != nil {
		return nil, err
	}

	// The type of network protocol and the full address (e.g., ":8989")
	network := "tcp"
	address := ":" + port

	// The object that waits for new incoming network connections
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	fmt.Println("Listening on the port", address)

	return listener, nil
}
