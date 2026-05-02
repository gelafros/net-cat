package internal

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Text   string
	Sender net.Conn
}

type Client struct {
	Name    string
	channel chan string
}

// Background writer goroutine dedicated to a single client
func clientWriter(conn net.Conn, ch chan string, name *string) {
	for msg := range ch {
		// Clear line, print the message, and reprint the prompt
		fmt.Fprintln(conn, "\r\033[2K"+msg)
		fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), *name)
	}
}

// Handles conns,names,chat history,and broadcasts names
func HandleClient(conn net.Conn, clients map[net.Conn]Client, key *sync.Mutex, broadcast chan Message, history *[]string) {
	var name string

	// Default color is standard terminal color (ANSI reset code)
	colorCode := "\033[0m"

	// Capture the client's network address for our server logs
	addr := conn.RemoteAddr().String()

	defer func() {
		// Waits until the whole func is run and safely removes the client from the clients map
		key.Lock()
		if clientStruct, ok := clients[conn]; ok && clientStruct.channel != nil {
			// Close the channel so the clientWriter goroutine stops cleanly
			close(clientStruct.channel)
		}
		delete(clients, conn)
		key.Unlock()

		// If the user successfully joined
		if name != "" {
			leaveMsg := fmt.Sprintf("%s has left our chat...", name)
			key.Lock()
			*history = append(*history, leaveMsg)
			key.Unlock()
			broadcast <- Message{Text: leaveMsg, Sender: conn}
		}
		// log.Printf("Client disconnected: %s\n", addr)
		log.Printf("User [%s] disconnected \n", name)
		conn.Close()
	}()

	// Welcome message and ASCII art logo
	logo := "Welcome to TCP-Chat!\n" +
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
	if _, err := fmt.Fprint(conn, logo); err != nil {
		log.Printf("Write error to %s: %v\n", addr, err)
		return
	}

	// Scanner to easily read text line-by-line from the clients conn
	msgscanner := bufio.NewScanner(conn)
	// Exit immediately if the client disconnects before entering anything
	if !msgscanner.Scan() {
		return
	}

	// The username chosen by the client
	name = msgscanner.Text()
	// Keep looping until a valid and unique name is provided
	for {
		if name == "" {
			if _, err := fmt.Fprintln(conn, "Invalid name, try again stupid"); err != nil {
				log.Printf("Write error to %s: %v\n", addr, err)
				return
			}
			fmt.Fprint(conn, "[ENTER YOUR NAME]: ")
		} else {
			// Safely check if the name is already used
			key.Lock()
			isTaken := false
			for _, existingClient := range clients {
				if existingClient.Name == name {
					isTaken = true
					break
				}
			}
			if !isTaken {
				// Initialize the struct and its dedicated channel
				clientChannel := make(chan string, 100) // Give them a mailbox that holds up to 100 unread messages
				clients[conn] = Client{Name: name, channel: clientChannel}

				log.Printf("Accepted connection for user: [%s]\n", name)
				key.Unlock()

				// Start the background writer specifically for this client
				go clientWriter(conn, clientChannel, &name)
				break
			}
			key.Unlock()
			if _, err := fmt.Fprintln(conn, "Name already taken, try another"); err != nil {
				log.Printf("Write error to %s: %v\n", addr, err)
				return
			}
			fmt.Fprint(conn, "[ENTER YOUR NAME]: ")
		}
		if !msgscanner.Scan() {
			return
		}
		name = msgscanner.Text()
	}

	// Show the available commands to the user
	instructions := "Available commands:\n" +
		"  --color=<color>  Change your message color (e.g., --color=red)\n" +
		"  --name=<name>    Change your display name (e.g., --name=Bob)\n" +
		"  --tableflip      Flip a table\n"
	if _, err := fmt.Fprintln(conn, instructions); err != nil {
		log.Printf("Write error to %s: %v\n", addr, err)
		return
	}

	// Print the past chat history for the new user
	key.Lock()
	for _, historyMsg := range *history {
		if _, err := fmt.Fprintln(conn, historyMsg); err != nil {
			log.Printf("Write error to %s: %v\n", addr, err)
			break
		}
	}
	key.Unlock()

	// Announce the new client to the rest of the chat room
	joinMsg := fmt.Sprintf("%s has joined our chat...", name)
	key.Lock()
	*history = append(*history, joinMsg)
	key.Unlock()

	// Send the join message to the central broadcast to send to receivers
	broadcast <- Message{Text: joinMsg, Sender: conn}

	// Print the initial typing prompt for the new user
	fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)

	// Wait for new messages from this client
	for msgscanner.Scan() {
		// The actual msg typed by the client
		msg := msgscanner.Text()

		// Move the cursor up and erase the raw input line from their terminal
		fmt.Fprint(conn, "\033[1A\r\033[2K")

		// Ignore empty lines to prevent chat spam
		if strings.TrimSpace(msg) == "" {
			// Redraw the prompt if they send an empty message
			fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user is trying to change their color
		if isColorCmd := handleColor(msg, conn, &colorCode, name, key, broadcast, history); isColorCmd {
			fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user is trying to change their name
		if isNameCmd := handleName(msg, conn, &name, clients, key, broadcast, history); isNameCmd {
			fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user wants to flip a table
		if strings.TrimSpace(msg) == "--tableflip" {
			msg = "(╯°□°)╯︵ ┻━┻"
		}

		// Format the message and send it to the central broadcast channel.
		formattedMsg := fmt.Sprintf("%s[%s][%s]:%s\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name, msg)

		// Echo the fully formatted and colored message back to the sender!
		fmt.Fprintln(conn, formattedMsg)

		key.Lock()
		*history = append(*history, formattedMsg)
		key.Unlock()
		broadcast <- Message{Text: formattedMsg, Sender: conn}

		// Redraw the prompt for the sender after they send a message
		fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)
	}

	// Log if the scanner stopped due to a network error instead of EOF
	if err := msgscanner.Err(); err != nil {
		log.Printf("Read error from %s: %v\n", addr, err)
	}
}
