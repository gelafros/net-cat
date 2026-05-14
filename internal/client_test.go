package internal

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestClientWriter(t *testing.T) {
	// Create an in-memory pipe to simulate a client connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	msgChan := make(chan string, 1)
	stateChan := make(chan clientState, 1)
	initialState := clientState{colorCode: "\033[0m", name: "TestUser"}

	// Start the clientWriter in the background
	go clientWriter(serverConn, msgChan, stateChan, initialState)

	// Send a test message through the channel
	msgChan <- "Hello Client"

	// Read the output from the simulated client connection
	buf := make([]byte, 1024)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	output := string(buf[:n])
	if !strings.Contains(output, "Hello Client") {
		t.Errorf("Expected output to contain 'Hello Client', got %q", output)
	}
}

func TestHandleClient(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	manager := NewRoomManager()

	// HandleClient expects TotalClients to be incremented before it starts (done in main.go)
	manager.TotalClients = 1

	done := make(chan bool)
	go func() {
		HandleClient(serverConn, manager)
		done <- true
	}()

	// We MUST continuously read from the client side of the pipe.
	// If we don't, HandleClient will block forever trying to write the massive ASCII logo!
	go func() {
		buf := make([]byte, 4096)
		for {
			if _, err := clientConn.Read(buf); err != nil {
				return
			}
		}
	}()

	// Give HandleClient a moment to write the logo and prompt
	time.Sleep(50 * time.Millisecond)

	// 1. Send a valid username
	if _, err := clientConn.Write([]byte("TestUser\n")); err != nil {
		t.Fatalf("Failed to write name: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify the user was added to the Main room
	room := manager.Rooms["Main"]
	room.Mutex.Lock()
	var found bool
	for _, c := range room.Clients {
		if c.Name == "TestUser" {
			found = true
			break
		}
	}
	room.Mutex.Unlock()
	if !found {
		t.Errorf("Expected TestUser to be joined to the Main room")
	}

	// 2. Send a normal chat message
	if _, err := clientConn.Write([]byte("Hello integration test\n")); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify the message made it into the room's history
	room.Mutex.Lock()
	history := room.History
	room.Mutex.Unlock()

	messageFound := false
	for _, msg := range history {
		if strings.Contains(msg, "Hello integration test") {
			messageFound = true
			break
		}
	}
	if !messageFound {
		t.Errorf("Expected chat history to contain 'Hello integration test'")
	}

	// 3. Disconnect and verify cleanup
	clientConn.Close()
	<-done // Wait for HandleClient to completely exit

	if manager.TotalClients != 0 {
		t.Errorf("Expected TotalClients to decrement to 0, got %d", manager.TotalClients)
	}
}
