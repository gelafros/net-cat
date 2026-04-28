package internal

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Handling the client(conn after listener.Accept is valid)
func HandleClient(conn net.Conn, clients map[net.Conn]bool, key *sync.Mutex, broadcast chan string, history *[]string) {
	var name string

	defer func() {
		key.Lock()
		delete(clients, conn)
		key.Unlock()

		if name != "" {
			leaveMsg := fmt.Sprintf("[%s] [%s] has left our chat...", time.Now().Format("2006-01-02 15:04:05"), name)
			key.Lock()
			*history = append(*history, leaveMsg)
			key.Unlock()
			broadcast <- leaveMsg
		}
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
	for _, historyMsg := range *history {
		fmt.Fprintln(conn, historyMsg)
	}
	key.Unlock()

	joinMsg := fmt.Sprintf("[%s] [%s] has joined our chat...", time.Now().Format("2006-01-02 15:04:05"), name)
	key.Lock()
	*history = append(*history, joinMsg)
	key.Unlock()
	broadcast <- joinMsg

	// Continuously wait for new messages from this client
	for msgscanner.Scan() {
		// The actual chat message typed by the client
		msg := msgscanner.Text()
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
}
