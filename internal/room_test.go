package internal

import (
	"net"
	"testing"
)

func TestRoomManager(t *testing.T) {
	rm := NewRoomManager()

	// Test that the Main room is created by default
	if _, exists := rm.Rooms["Main"]; !exists {
		t.Errorf("Expected Main room to exist upon initialization")
	}

	// Test CreateRoom
	err := rm.CreateRoom("Gaming")
	if err != nil {
		t.Errorf("Expected no error when creating room, got %v", err)
	}

	if _, exists := rm.Rooms["Gaming"]; !exists {
		t.Errorf("Expected Gaming room to be present in Rooms map")
	}

	// Test duplicate room creation
	err = rm.CreateRoom("Gaming")
	if err == nil {
		t.Errorf("Expected an error when creating a room that already exists")
	}

	// Test ListRooms
	rooms := rm.ListRooms()
	if len(rooms) != 2 {
		t.Errorf("Expected 2 active rooms (Main and Gaming), got %d", len(rooms))
	}
}

func TestChangeRoomAndRemoveClient(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("Gaming")

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Drain the output so writing to the connection doesn't block the test
	go func() {
		buf := make([]byte, 1024)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	mainRoom := rm.Rooms["Main"]
	mainRoom.Clients[clientConn] = Client{Name: "TestUser", channel: make(chan string, 10)}

	// Test ChangeRoom (Success)
	targetRoom, err := rm.ChangeRoom(mainRoom, "Gaming", clientConn, "TestUser")
	if err != nil {
		t.Errorf("Unexpected error changing room: %v", err)
	}
	if _, exists := mainRoom.Clients[clientConn]; exists {
		t.Errorf("Expected client to be removed from Main room")
	}
	if _, exists := targetRoom.Clients[clientConn]; !exists {
		t.Errorf("Expected client to be added to Gaming room")
	}

	// Test ChangeRoom (Fail: Room doesn't exist)
	if _, err = rm.ChangeRoom(targetRoom, "NonExistent", clientConn, "TestUser"); err == nil {
		t.Errorf("Expected error when joining non-existent room")
	}

	// Test RemoveClient
	rm.RemoveClient(targetRoom, clientConn, "TestUser")
	if len(targetRoom.Clients) != 0 {
		t.Errorf("Expected target room to be empty after RemoveClient")
	}
}