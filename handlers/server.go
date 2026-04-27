package handlers

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	mu         sync.RWMutex
	clientList map[net.Conn]string
	broadcast  chan Message
}

// Pointer to type *Server
func NewServer() *Server {
	return &Server{
		clientList: make(map[net.Conn]string),
		broadcast:  make(chan Message),
	}
}

// Broadcasting Client's message func
func (s *Server) StartBroadcastLoop() {
	for m := range s.broadcast {
		// fmt.Printf("[BROADCAST] [%s]: %s\n", m.Name, m.Content) // Broadcast message to Server
		s.mu.RLock() //use RLock (read-only), not Lock
		for client := range s.clientList {
			if client == m.Sender {
				continue // skip current client (the one who broadcast)
			}
			client.Write([]byte(fmt.Sprintf("[%s]: %s\n", m.Name, m.Content)))
		}
		s.mu.RUnlock() // use RUnlock (read-only), not Unlock
	}
}

// Adding Client with Server secure from concurrent access with Mutex
func (s *Server) TryToAddClient(client net.Conn, name string) bool {
	s.mu.Lock()         // lock the goroutine and waits until finished
	defer s.mu.Unlock() // unlock goroutine when is complete

	if len(s.clientList) >= 10 {
		// client.Write([]byte("Server is Full!"))
		return false
	}

	for _, n := range s.clientList {
		if n == name {
			return false
		}
	}
	s.clientList[client] = name // register Client's name to clientList
	return true
}

// Removes a client (User) from client list securely (with mu.Lock)
func (s *Server) RemoveClient(client net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clientList, client)
}

// Return the Number of Users (securely with mu.Lock)
// func (s *Server) ClientCount() int {
// 	s.mu.RLock()         //use RLock (read-only), not Lock
// 	defer s.mu.RUnlock() // use RUnlock (read-only), not Unlock
// 	return len(s.clientList)
// }
