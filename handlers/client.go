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
func handshake(client net.Conn, server *Server, reader *bufio.Reader) (string, error) {
	var validName string
	var validError error
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

	// Loop for accepting User's input
	for {
		userInput, err := userMessage(reader)
		if err != nil {
			fmt.Println("Failed to read username:", err)
			validError = err
			client.Close()
			break
		}
		if userInput == "" {
			client.Write([]byte("You must enter a name to access the chat room.\n[ENTER YOUR NAME]: "))
			continue // Keep the loop
		}

		if !server.TryToAddClient(client, userInput) {
			client.Write([]byte("Name in use or server full.\n[ENTER YOUR NAME]: "))
			continue
		}

		//Accept User (if not two above conditions are true)
		message := strings.TrimSpace("Welcome to chat room, " + userInput)
		client.Write([]byte(message)) // Message to Client
		client.Write([]byte("\n"))
		message = strings.TrimSpace("User [" + userInput + "] has entered the Chat room")
		fmt.Println(message) // Message to Server
		validName = userInput
		validError = nil
		break // End of the "Acception" loop
	}
	return validName, validError
}

func userMessage(reader *bufio.Reader) (string, error) {
	var msg string
	msg, err := reader.ReadString(('\n'))
	if err != nil {
		// fmt.Println("Failed in userMessage:", err)
		return "", err
	}
	msg = strings.TrimSpace(msg)
	return msg, err
}

// Handling EACH client that connects to the server
func HandleClient(client net.Conn, server *Server) {
	var clName string

	reader := bufio.NewReader(client)                    // only ONE reader for this client
	clientName, err := handshake(client, server, reader) // User login with name

	if err != nil {
		client.Close()
		return
	}
	clName = clientName

	// Waiting for Client input loop
	for {
		message, err := userMessage(reader) // Re-use same reader
		if err != nil {
			// delete(server.clientList, client)
			client.Close()              // disconnect client
			server.RemoveClient(client) // Use server method RemoveClient(client) for securely remove client
			fmt.Println(strings.TrimSpace("User [" + clName + "] has left the Chat room"))
			break
		}
		if message == "" {
			continue
		}

		// Register userMessage to server.broadcast channel
		server.broadcast <- Message{
			Sender:  client,
			Name:    clName,
			Content: message,
		}
	}
}
