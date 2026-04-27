package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
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

	users := 0
	var key sync.Mutex

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		} else {
			key.Lock()
			if users == 10 {
				fmt.Fprintln(conn, "Server busy")
				conn.Close()
				key.Unlock()
				continue
			} else {
				users++
				key.Unlock()
			}
		}

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
