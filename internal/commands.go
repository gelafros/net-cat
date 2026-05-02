package internal

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// Checks if the user typed a color command and returns the color code
func handleColor(msg string, conn net.Conn, colorCode *string, name string, key *sync.Mutex, broadcast chan Message, history *[]string) bool {
	if !strings.HasPrefix(msg, "--color") {
		return false
	}

	// Remove "--color" and "=" to get just the color name
	colorInput := strings.TrimSpace(strings.TrimPrefix(msg, "--color"))
	colorInput = strings.TrimPrefix(colorInput, "=")
	colorInput = strings.TrimSpace(colorInput)

	var newColor string
	switch strings.ToLower(colorInput) {
	case "red":
		newColor = "\033[31m"
	case "green":
		newColor = "\033[32m"
	case "yellow":
		newColor = "\033[33m"
	case "blue":
		newColor = "\033[34m"
	case "magenta":
		newColor = "\033[35m"
	case "cyan":
		newColor = "\033[36m"
	case "reset":
		newColor = "\033[0m"
	default:
		fmt.Fprintln(conn, "Unknown color. Try: red, green, yellow, blue, magenta, cyan, reset")
		return true
	}

	*colorCode = newColor
	changeMsg := fmt.Sprintf("User [%s] has changed their color to %s", name, colorInput)

	key.Lock()
	*history = append(*history, changeMsg)
	key.Unlock()

	fmt.Fprintln(conn, changeMsg)                       // Echo to sender
	broadcast <- Message{Text: changeMsg, Sender: conn} // Broadcast to others

	return true
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

// Checks if the user requested the help menu
func HandleHelp(msg string, conn net.Conn) bool {
	if strings.TrimSpace(msg) == "/help" {
		instructions := "Available commands:\n" +
			"  --color=<color>  Change your message color (e.g., --color=red)\n" +
			"  --name=<name>    Change your display name (e.g., --name=Bob)\n" +
			"  --tableflip      Flip a table\n" +
			"  /help            Show this help menu"
		fmt.Fprintln(conn, instructions)
		return true
	}
	return false
}

// Transforms the tableflip command into the ASCII art
func HandleTableFlip(msg string) string {
	if strings.TrimSpace(msg) == "--tableflip" {
		return "(╯°□°)╯︵ ┻━┻"
	}
	return msg
}
