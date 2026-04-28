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
func HandleClient(conn net.Conn, clients map[net.Conn]bool, key *sync.Mutex, broadcast chan string) {
	var name string

	defer func() {
		key.Lock()
		delete(clients, conn)
		key.Unlock()

		if name != "" {
			broadcast <- fmt.Sprintf("[%s] [%s] has left our chat...", time.Now().Format("2006-01-02 15:04:05"), name)
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
	broadcast <- fmt.Sprintf("[%s] [%s] has joined our chat...", time.Now().Format("2006-01-02 15:04:05"), name)

	// Continuously wait for new messages from this client
	for msgscanner.Scan() {
		// The actual chat message typed by the client
		msg := msgscanner.Text()
		if strings.TrimSpace(msg) == "" {
			continue
		}
		// Format the message and send it to the central broadcast channel.
		formattedMsg := fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), name, msg)
		broadcast <- formattedMsg
	}
}
