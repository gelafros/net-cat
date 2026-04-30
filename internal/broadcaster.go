package internal

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Background goroutine that listens for messages and sends them to all clients
func Broadcaster(broadcast chan Message, clients map[net.Conn]string, key *sync.Mutex) {
	for {
		// Pause and wait for a new message to be sent into the channel
		msg := <-broadcast

		key.Lock()
		// Send the message to every connected client in the chat room
		for receiver, receiverName := range clients {
			// Only send to clients who have joined (have a name) and aren't the sender
			if receiver != msg.Sender && receiverName != "" {
				if _, err := fmt.Fprintln(receiver, "\r\033[2K"+msg.Text); err != nil {
					log.Printf("Error sending message to %s: %v\n", receiver.RemoteAddr().String(), err)
				}
				fmt.Fprintf(receiver, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), receiverName)
			}
		}
		key.Unlock()
	}
}
