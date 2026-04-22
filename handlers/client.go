package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Handshake Client with Server ("Continue" for invalid (empty name) / "Break" for valid names )
func handshake(client net.Conn) string {
	var validName string
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
	client.Write([]byte(chatLogo))
	for {
		reader := bufio.NewReader(client)
		name, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read username:", err)
			client.Close()
			break
		}
		name = strings.TrimSpace(name) // Clean name from spaces (trim)
		if name == "" {
			client.Write([]byte("You must enter a name to access the chat room.\n[ENTER YOUR NAME]: "))
			continue // Keep the loop
		} else {
			validName = name
			break // End the loop after a given valid name
		}
	}
	return validName
}

// Handling EACH client that connects to the server

func HandleClient(client net.Conn, server *Server) {
	// const chatLogo = "Welcome to TCP-Chat!\n" +
	// 	"         _nnnn_\n" +
	// 	"        dGGGGMMb\n" +
	// 	"       @p~qp~~qMb\n" +
	// 	"       M|@||@) M|\n" +
	// 	"       @,----.JM|\n" +
	// 	"      JS^\\__/  qKL\n" +
	// 	"     dZP        qKRb\n" +
	// 	"    dZP          qKKb\n" +
	// 	"   fZP            SMMb\n" +
	// 	"   HZM            MMMM\n" +
	// 	"   FqM            MMMM\n" +
	// 	" __| \".        |\\dS\"qML\n" +
	// 	" |    `.       | `' \\Zq\n" +
	// 	"_)      \\.___.,|     .'\n" +
	// 	"\\____   )MMMMMP|   .'\n" +
	// 	"     `-'       `--'\n" +
	// 	"[ENTER YOUR NAME]: "
	var clientName string

	// if !maxCapacity {
	// 	fmt.Println("A new user has joined:", client.RemoteAddr())
	// 	client.Write([]byte(serverMsg))

	// Client validation with a name
	// client.Write([]byte(chatLogo))
	clientName = handshake(client)
	server.clientList[client] = clientName
	message := strings.TrimSpace("Welcome to chat room, " + clientName)
	client.Write([]byte(message)) // Message to Client
	client.Write([]byte("\n"))
	server.clientList[client] = clientName // register Client's name to list
	message = strings.TrimSpace("User \"" + clientName + "\" has entered the Chat Room")
	fmt.Println(message) // Message to Server
}
