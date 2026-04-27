package main

import (
	"errors"
	"fmt"
	"net"
	"os"
)

func main() {
	args := os.Args
	port, err := getPort(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	network := "tcp"
	address := ":" + port

	listener, err := net.Listen(network, address)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
		return
	}

	fmt.Println("Listening on the port", address)
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
