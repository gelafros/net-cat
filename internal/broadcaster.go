package internal

import (
	"net"
	"sync"
)

// Background goroutine that listens for messages and sends them to all clients
func Broadcaster(broadcast chan Message, clients map[net.Conn]Client, key *sync.Mutex) {
	for {
		// Pause and wait for a new message to be sent into the channel
		msg := <-broadcast

		key.Lock()
		// Send the message to every connected client in the chat room
		for receiver, clientInfo := range clients {
			// Only send to clients who have joined (have a name) and aren't the sender
			if receiver != msg.Sender && clientInfo.Name != "" {
				// Instantly send message to the channel in memory; no network blocking!
				clientInfo.channel <- msg.Text
			}
		}
		key.Unlock()
	}
}
