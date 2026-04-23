package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Message struct {
	Sender  net.Conn
	Name    string
	Content string
}

// Handshake Client with Server ("Continue" for invalid (empty name) / "Break" for valid names )
func handshake(client net.Conn, userInput string, err error) string {
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
	// reader := bufio.NewReader(client)
	for {
		// name, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read username:", err)
			client.Close()
			break
		}
		// name = strings.TrimSpace(name) // Clean name from spaces (trim)
		if userInput == "" {
			client.Write([]byte("You must enter a name to access the chat room.\n[ENTER YOUR NAME]: "))
			continue // Keep the loop
		} else {
			validName = userInput
			break // End the loop after a given valid name
		}
	}
	return validName
}

func userMessage(client net.Conn) (string, error) {
	var msg string
	reader := bufio.NewReader(client)
	msg, err := reader.ReadString(('\n'))
	if err != nil {
		return "", err
	}
	msg = strings.TrimSpace(msg)
	return msg, nil
}

// Handling EACH client that connects to the server

func HandleClient(client net.Conn, server *Server) {
	var clientName string
	// reader := bufio.NewReader(client)
	// msg, err := reader.ReadString(('\n'))
	// if err != nil {
	// 	fmt.Println("Error in reading User input:", err)
	// } else {
	// 	err = nil
	// }
	// msg = strings.TrimSpace(msg)
	msg := userMessage(client)
	clientName = handshake(client, msg, err)

	server.clientList[client] = clientName
	message := strings.TrimSpace("Welcome to chat room, " + clientName)
	client.Write([]byte(message)) // Message to Client
	client.Write([]byte("\n"))
	server.clientList[client] = clientName // register Client's name to list
	message = strings.TrimSpace("User \"" + clientName + "\" has entered the Chat Room")
	fmt.Println(message) // Message to Server
	fmt.Println("Client list from Server:", server.clientList)

	for {
		message, err := userMessage(client)
		if err != nil {
			delete(server.clientList, client)
			client.Close()
			fmt.Println(strings.TrimSpace("User \"" + clientName + "\" disconnected"))
			fmt.Println("Client list from Server:", server.clientList)
			// return
			break
		}
		if message == "" {
			continue
		}
		// fmt.Println(clientName, " said:", message)
		server.broadcast <- Message{
			Sender:  client,
			Name:    clientName,
			Content: message,
		}
	}

	// server.broadcast[message] = readInput(client)
}
