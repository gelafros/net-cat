package handlers

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

// Handling the client(conn after listener.Accept is valid)
func HandleClient(conn net.Conn, clients map[net.Conn]bool, key *sync.Mutex) {
	var name string

	defer func() {
		key.Lock()
		delete(clients, conn)
		if name != "" {
			for receiver := range clients {
				fmt.Fprintln(receiver, "["+time.Now().Format("2006-01-02 15:04:05")+"]", "["+name+"]", ":", "has left our chat...")
			}
		}
		key.Unlock()
		conn.Close()
	}()

	fmt.Fprintln(conn, "[ENTER YOUR NAME]")
	// Scanner to easily read text line-by-line from the client's connection
	msgscanner := bufio.NewScanner(conn)
	if msgscanner.Scan() == false {
		return
	}

	// The username chosen by the client
	name = msgscanner.Text()
	for name == "" {
		fmt.Fprintln(conn, "Invalid name, try again")
		if msgscanner.Scan() == false {
			return
		}
		name = msgscanner.Text()
	}

	fmt.Fprintln(conn, "Welcome", name)
	key.Lock()
	// Announce the new user to everyone else (receivers) in the chat
	for receiver := range clients {
		if receiver != conn {
			fmt.Fprintln(receiver, "["+time.Now().Format("2006-01-02 15:04:05")+"]", "["+name+"]", ":", "has joined our chat...")
		}
	}
	key.Unlock()

	// Continuously wait for new messages from this client
	for msgscanner.Scan() {
		// The actual chat message typed by the client
		msg := msgscanner.Text()
		// Format the message once to ensure a consistent timestamp for all clients
		formattedMsg := fmt.Sprintf("[%s] [%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, msg)
		key.Lock()
		for receiver := range clients {
			fmt.Fprintln(receiver, formattedMsg)
		}
		key.Unlock()
	}
}
