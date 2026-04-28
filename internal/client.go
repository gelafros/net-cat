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

// Handles conns,names,chat history,and broadcasts names
func HandleClient(conn net.Conn, clients map[net.Conn]string, key *sync.Mutex, broadcast chan Message, history *[]string) {
	var name string

	// Default color is standard terminal color (ANSI reset code)
	colorCode := "\033[0m"

	// Capture the client's network address for our server logs
	addr := conn.RemoteAddr().String()

	defer func() {
		// Waits until the whole func is run and safely removes the client from the clients map
		key.Lock()
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
		log.Printf("Client disconnected: %s\n", addr)
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
	if msgscanner.Scan() == false {
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
		} else {
			key.Lock()
			isTaken := false
			for _, existingName := range clients {
				if existingName == name {
					isTaken = true
					break
				}
			}
			if !isTaken {
				clients[conn] = name // Claim the name while still locked
				key.Unlock()
				break
			}
			key.Unlock()
			if _, err := fmt.Fprintln(conn, "Name already taken, try another"); err != nil {
				log.Printf("Write error to %s: %v\n", addr, err)
				return
			}
		}
		if msgscanner.Scan() == false {
			return
		}
		name = msgscanner.Text()
	}

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
		if newColor, isColorCmd := handleColor(msg, conn); isColorCmd {
			if newColor != "" {
				colorCode = newColor
			}
			// Redraw the prompt and skip broadcasting this command
			fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user is trying to change their name
		if isNameCmd := handleName(msg, conn, &name, clients, key, broadcast, history); isNameCmd {
			fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)
			continue
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

// handleColorCmd parses a color command. Returns the ANSI color string and a boolean indicating if it was a color command.
func handleColor(msg string, conn net.Conn) (string, bool) {
	if !strings.HasPrefix(msg, "--color") {
		return "", false
	}

	colorInput := strings.TrimSpace(strings.TrimPrefix(msg, "--color"))
	colorInput = strings.TrimPrefix(colorInput, "=")
	colorInput = strings.TrimSpace(colorInput)

	switch strings.ToLower(colorInput) {
	case "red":
		return "\033[31m", true
	case "green":
		return "\033[32m", true
	case "yellow":
		return "\033[33m", true
	case "blue":
		return "\033[34m", true
	case "magenta":
		return "\033[35m", true
	case "cyan":
		return "\033[36m", true
	case "reset":
		return "\033[0m", true
	default:
		fmt.Fprintln(conn, "Unknown color. Try: red, green, yellow, blue, magenta, cyan, reset")
		return "", true
	}
}

// handleNameCmd updates the user's name across the server state if a name command is issued.
func handleName(msg string, conn net.Conn, name *string, clients map[net.Conn]string, key *sync.Mutex, broadcast chan Message, history *[]string) bool {
	if !strings.HasPrefix(msg, "--name") {
		return false
	}

	newName := strings.TrimSpace(strings.TrimPrefix(msg, "--name"))
	newName = strings.TrimPrefix(newName, "=")
	newName = strings.TrimSpace(newName)

	if newName == "" {
		fmt.Fprintln(conn, "Name cannot be empty. Try: --name=NewName")
		return true
	}

	key.Lock()
	isTaken := false
	for _, existingName := range clients {
		if existingName == newName {
			isTaken = true
			break
		}
	}
	if isTaken {
		key.Unlock()
		fmt.Fprintln(conn, "Name already in use. Try a different name.")
		return true
	}

	oldName := *name
	*name = newName // Update new name

	clients[conn] = *name // Update name in the central map
	changeMsg := fmt.Sprintf("%s changed their name to %s", oldName, *name)
	*history = append(*history, changeMsg)
	key.Unlock()

	fmt.Fprintln(conn, changeMsg)                       // Echo to sender
	broadcast <- Message{Text: changeMsg, Sender: conn} // Broadcast to others

	return true
}
