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

// Handles conns,names,chat history,and broadcasts names
func HandleClient(conn net.Conn, clients map[net.Conn]bool, key *sync.Mutex, broadcast chan string, history *[]string) {
	var name string

	// Capture the client's network address for our server logs
	addr := conn.RemoteAddr().String()

	defer func() {
		// Waits until the whole func is run and safely removes the client from the clients map
		key.Lock()
		delete(clients, conn)
		key.Unlock()

		// If the user successfully joined
		if name != "" {
			leaveMsg := fmt.Sprintf("[%s] [%s] has left our chat...", time.Now().Format("2006-01-02 15:04:05"), name)
			key.Lock()
			*history = append(*history, leaveMsg)
			key.Unlock()
			broadcast <- leaveMsg
		}
		log.Printf("Client disconnected: %s\n", addr)
		conn.Close()
	}()

	// Make new users to enter their name
	if _, err := fmt.Fprintln(conn, "[ENTER YOUR NAME]"); err != nil {
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
	// Keep looping until a non-empty name is provided
	for name == "" {
		if _, err := fmt.Fprintln(conn, "Invalid name, try again"); err != nil {
			log.Printf("Write error to %s: %v\n", addr, err)
			return
		}
		if msgscanner.Scan() == false {
			return
		}
		name = msgscanner.Text()
	}

	if _, err := fmt.Fprintln(conn, "Welcome", name); err != nil {
		log.Printf("Write error to %s: %v\n", addr, err)
		return
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
	joinMsg := fmt.Sprintf("[%s] [%s] has joined our chat...", time.Now().Format("2006-01-02 15:04:05"), name)
	key.Lock()
	*history = append(*history, joinMsg)
	key.Unlock()

	// Send the join message to the central broadcast to send to receivers
	broadcast <- joinMsg

	// Wait for new messages from this client
	for msgscanner.Scan() {
		// The actual msg typed by the client
		msg := msgscanner.Text()
		// Ignore empty lines to prevent chat spam
		if strings.TrimSpace(msg) == "" {
			continue
		}
		// Format the message and send it to the central broadcast channel.
		formattedMsg := fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), name, msg)
		key.Lock()
		*history = append(*history, formattedMsg)
		key.Unlock()
		broadcast <- formattedMsg
	}

	// Log if the scanner stopped due to a network error instead of EOF
	if err := msgscanner.Err(); err != nil {
		log.Printf("Read error from %s: %v\n", addr, err)
	}
}
