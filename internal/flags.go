package internal

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// Checks if the user typed a color command and returns the color code
func handleColor(msg string, conn net.Conn) (string, bool) {
	if !strings.HasPrefix(msg, "--color") {
		return "", false
	}

	// Remove "--color" and "=" to get just the color name
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

// Changes the user's name if they use the name command and the name is free
func handleName(msg string, conn net.Conn, name *string, clients map[net.Conn]Client, key *sync.Mutex, broadcast chan Message, history *[]string) bool {
	if !strings.HasPrefix(msg, "--name") {
		return false
	}

	// Remove "--name" and "=" to get the new name
	newName := strings.TrimSpace(strings.TrimPrefix(msg, "--name"))
	newName = strings.TrimPrefix(newName, "=")
	newName = strings.TrimSpace(newName)

	if newName == "" {
		fmt.Fprintln(conn, "Name cannot be empty. Try: --name=NewName")
		return true
	}

	// Check if the new name is already taken
	key.Lock()
	isTaken := false
	for _, existingClient := range clients {
		if existingClient.Name == newName {
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

	// Pull the struct out, update the name, and put it back
	c := clients[conn]
	c.Name = *name
	clients[conn] = c
	changeMsg := fmt.Sprintf("[%s] changed their name to [%s]", oldName, *name)
	*history = append(*history, changeMsg)
	key.Unlock()

	fmt.Fprintln(conn, changeMsg)                       // Echo to sender
	broadcast <- Message{Text: changeMsg, Sender: conn} // Broadcast to others

	return true
}
