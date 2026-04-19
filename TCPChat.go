package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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

// Handling EACH client that connects to the server
func handleClient(client net.Conn, serverMsg string, maxCapacity bool) {

	if !maxCapacity {
		fmt.Println("A new user has joined:", client.RemoteAddr())
		client.Write([]byte(serverMsg))
	} else { // Refuse client's connection
		fmt.Println("The user: ", client.RemoteAddr(), " has been rejected due to Full Server")
		client.Write([]byte(serverMsg))
		client.Close()
	}
}

func main() {
	args := os.Args
	clientCnt := 0
	ServerIsFull := false
	srvMsg := strings.TrimSpace("Welcome to TCP-Chat!\n+
	 _nnnn_\n+
        dGGGGMMb\n+
       @p~qp~~qMb\n+
       M|@||@) M|\n+
       @,----.JM|\n+
      JS^\__/  qKL\n+
     dZP        qKRb\n+
    dZP          qKKb\n+
   fZP            SMMb\n+
   HZM            MMMM\n+
   FqM            MMMM\n+
 __| ".        |\dS"qML\n+
 |    `.       | `' \Zq\n+
_)      \.___.,|     .'\n+
\____   )MMMMMP|   .'\n+
     `-'       `--'\n+
[ENTER YOUR NAME]:")
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

	// Running loop for receiving multiple clients
	for {
		client, err := listener.Accept()
		if err != nil {
			fmt.Println("Error with client connection:", err)
			continue // for not terminating the Server
		}
		clientCnt++

		// Checking MAX Server capacity
		if clientCnt > 10 {
			ServerIsFull = true
			srvMsg = "Server Busy"
		}
		handleClient(client, srvMsg, ServerIsFull)
	}
}
