package internal

import (
	"errors"
	"log"
	"net"
)

// Decides which port we are listening on. Default 8989
func getPort(args []string) (string, error) {
	if len(args) == 1 {
		return "8989", nil
	}

	if len(args) == 2 {
		return args[1], nil
	}

	return "", errors.New("[USAGE]: ./TCPChat $port")
}

// StartServer starts a TCP listener
func StartServer(args []string) (net.Listener, error) {
	// The port number the server will listen on
	port, err := getPort(args)
	if err != nil {
		return nil, err
	}

	// The type of protocol and the address(port)
	network := "tcp"
	address := ":" + port

	// The object that waits for connections
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	log.Println("Listening on the port", address)

	return listener, nil
}
