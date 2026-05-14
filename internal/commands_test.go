package internal

import (
	"net"
	"sync"
	"testing"
)

func TestHandleTableFlip(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid command", "/tableflip", "(╯°□°)╯︵ ┻━┻"},
		{"Valid with extra spaces", " /tableflip ", "(╯°□°)╯︵ ┻━┻"},
		{"Normal message", "Hello world", "Hello world"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleTableFlip(tt.input)
			if got != tt.expected {
				t.Errorf("HandleTableFlip(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHandleColor(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	colorCode := "\033[0m"
	history := make([]string, 0)
	var key sync.Mutex
	broadcast := make(chan Message, 10)

	// Valid color command
	handled := handleColor("/color=red", clientConn, &colorCode, "TestUser", &key, broadcast, &history)
	if !handled {
		t.Errorf("Expected handleColor to process a valid color command")
	}
	if colorCode != "\033[31m" {
		t.Errorf("Expected colorCode to change to red ANSI code, got %q", colorCode)
	}

	// Invalid command handling
	handled = handleColor("/name=Bob", clientConn, &colorCode, "TestUser", &key, broadcast, &history)
	if handled {
		t.Errorf("Expected handleColor to ignore non-color commands")
	}
}

func TestHandleHelp(t *testing.T) {
	// net.Pipe() gives us two ends of a fake connection.
	// What we write to clientConn comes out of serverConn, and vice versa.
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// We need a background routine to continuously read from the serverConn.
	// If we don't do this, HandleHelp will freeze forever waiting for someone to read its output!
	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid help command", "/help", true},
		{"Not a help command", "/name=Bob", false},
		{"Regular text", "I need help", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandleHelp(tt.input, clientConn); got != tt.expected {
				t.Errorf("HandleHelp(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHandleName(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	name := "OldName"
	var key sync.Mutex
	broadcast := make(chan Message, 10)
	history := make([]string, 0)
	clients := make(map[net.Conn]Client)
	clients[clientConn] = Client{Name: "OldName", channel: make(chan string, 10)}

	// Setup another user to test duplicate names
	otherConn, otherServer := net.Pipe()
	defer otherConn.Close()
	defer otherServer.Close()
	clients[otherConn] = Client{Name: "Bob", channel: make(chan string, 10)}

	// Test Valid Name Change
	handled := handleName("/name=NewName", clientConn, &name, clients, &key, broadcast, &history)
	if !handled || name != "NewName" {
		t.Errorf("Expected valid name change to process and update name, got handled=%v, name=%s", handled, name)
	}

	// Test Duplicate Name Change
	handled = handleName("/name=Bob", clientConn, &name, clients, &key, broadcast, &history)
	if !handled || name == "Bob" {
		t.Errorf("Expected duplicate name to be rejected, got handled=%v, name=%s", handled, name)
	}

	// Test Non-name command
	if handled = handleName("/help", clientConn, &name, clients, &key, broadcast, &history); handled {
		t.Errorf("Expected non-name command to be ignored")
	}
}

func TestHandleWhisper(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	var key sync.Mutex
	clients := make(map[net.Conn]Client)
	clients[clientConn] = Client{Name: "Sender", channel: make(chan string, 10)}

	otherConn, otherServer := net.Pipe()
	defer otherConn.Close()
	defer otherServer.Close()
	targetChan := make(chan string, 10)
	clients[otherConn] = Client{Name: "Target", channel: targetChan}

	// Test Valid Whisper
	if handled := handleWhisper("/whisper=Target hello", clientConn, "Sender", clients, &key); !handled {
		t.Errorf("Expected valid whisper to be handled")
	}

	// Test Whisper to Self
	if handled := handleWhisper("/whisper=Sender hi", clientConn, "Sender", clients, &key); !handled {
		t.Errorf("Expected whisper to self to be handled (rejected but handled=true)")
	}

	// Test Invalid Command
	if handled := handleWhisper("/color=red", clientConn, "Sender", clients, &key); handled {
		t.Errorf("Expected non-whisper command to be ignored")
	}
}

func TestHandleRooms(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	manager := NewRoomManager()

	// Valid command
	if handled := handleRooms("/rooms", clientConn, manager); !handled {
		t.Errorf("Expected /rooms to be handled")
	}

	// Invalid command
	if handled := handleRooms("/help", clientConn, manager); handled {
		t.Errorf("Expected non-rooms command to be ignored")
	}
}

func TestHandleCreate(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	manager := NewRoomManager()

	// Valid create
	if handled := handleCreate("/create=TestRoom", clientConn, manager); !handled {
		t.Errorf("Expected /create to be handled")
	}
	if _, exists := manager.Rooms["TestRoom"]; !exists {
		t.Errorf("Expected TestRoom to be created")
	}

	// Invalid create (empty name)
	if handled := handleCreate("/create=", clientConn, manager); !handled {
		t.Errorf("Expected empty create to be handled (and rejected)")
	}

	// Invalid command
	if handled := handleCreate("/help", clientConn, manager); handled {
		t.Errorf("Expected non-create command to be ignored")
	}
}

func TestHandleJoin(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	manager := NewRoomManager()
	manager.CreateRoom("TestRoom")

	currRoom := manager.Rooms["Main"]
	currRoom.Clients[clientConn] = Client{Name: "Tester", channel: make(chan string, 10)}

	// Valid join
	if handled := handleJoin("/join=TestRoom", clientConn, "Tester", &currRoom, manager); !handled {
		t.Errorf("Expected /join to be handled")
	}
	if currRoom.Name != "TestRoom" {
		t.Errorf("Expected currRoom to change to TestRoom")
	}

	// Invalid command
	if handled := handleJoin("/help", clientConn, "Tester", &currRoom, manager); handled {
		t.Errorf("Expected non-join command to be ignored")
	}
}

func TestHandleLeave(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, err := serverConn.Read(buf); err != nil {
				return
			}
		}
	}()

	manager := NewRoomManager()
	manager.CreateRoom("TestRoom")

	testRoom := manager.Rooms["TestRoom"]

	currRoom := testRoom
	testRoom.Clients[clientConn] = Client{Name: "Tester", channel: make(chan string, 10)}

	// Valid leave
	if handled := handleLeave("/leave", clientConn, "Tester", &currRoom, manager); !handled {
		t.Errorf("Expected /leave to be handled")
	}
	if currRoom.Name != "Main" {
		t.Errorf("Expected currRoom to change to Main")
	}

	// Leave while in Main
	if handled := handleLeave("/leave", clientConn, "Tester", &currRoom, manager); !handled {
		t.Errorf("Expected /leave while in Main to be handled")
	}

	// Invalid command
	if handled := handleLeave("/help", clientConn, "Tester", &currRoom, manager); handled {
		t.Errorf("Expected non-leave command to be ignored")
	}
}
