package main

import (
	"bufio"
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
		var name string
		fmt.Println("A new user has joined:", client.RemoteAddr())
		client.Write([]byte(serverMsg))
		reader := bufio.NewReader(client)
		// name, _ := reader.ReadString('\n')
		// name = strings.TrimSpace(name)

		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Failed to read username:", err)
				client.Close()
				return
			}

			name = strings.TrimSpace(input)
			// if name != "" {
			// 	break
			// }

			if name != "" {
				message := strings.TrimSpace("Welcome to chat room, " + name + "\n")
				client.Write([]byte(message + "\n"))
				// if _, err := client.Write([]byte("You must enter a name to access the chat room.\n[ENTER YOUR NAME]: ")); err != nil {
				// 	fmt.Println("Failed to re-prompt username:", err)
				// 	client.Close()
				// 	return
				// }
			}

			// message := strings.TrimSpace("Welcome to chat room, " + name + "\n")
			// client.Write([]byte(message))

			// if name != "" {
			// 	message := strings.TrimSpace("Welcome to chat room, " + name + "\n")
			// 	client.Write([]byte(message + "\n"))

			// 	// Message to Server with Print
			// 	message = strings.TrimSpace("User \"" + name + "\" has enter the Chat Room!")
			// 	fmt.Println(message)
			// } else {
			// 	client.Write([]byte("You must enter a Name for access the Chat room"))
			// 	client.Write([]byte("[ENTER YOUR NAME]: "))
			// 	client.Close()
			// }

		}
	}
}

func main() {
	const chatLogo = "Welcome to TCP-Chat!\n" +
		"         _nnnn_\n" +
		"        dGGGGMMb\n" +
		"       @p~qp~~qMb\n" +
		"       M|@||@) M|\n" +
		"       @,----.JM|\n" +
		"      JS^\\__/  qKL\n" +
		"     dZP        qKRb\n" +
		"    dZP          qKKb\n" +
		"   fZP            SMMb\n" +
		"   HZM            MMMM\n" +
		"   FqM            MMMM\n" +
		" __| \".        |\\dS\"qML\n" +
		" |    `.       | `' \\Zq\n" +
		"_)      \\.___.,|     .'\n" +
		"\\____   )MMMMMP|   .'\n" +
		"     `-'       `--'\n" +
		"[ENTER YOUR NAME]: "
	args := os.Args
	clientCnt := 0
	ServerIsFull := false
	srvMsg := chatLogo
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
