package internal

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// Room represents a standalone chat room with its own state and broadcaster.
type Room struct {
	Name      string
	Clients   map[net.Conn]Client
	History   []string
	Broadcast chan Message
	Mutex     sync.Mutex
}

// NewRoom initializes and returns a new Room.
func NewRoom(name string) *Room {
	return &Room{
		Name:      name,
		Clients:   make(map[net.Conn]Client),
		Broadcast: make(chan Message),
	}
}

// RoomManager tracks multiple chat rooms safely.
type RoomManager struct {
	Rooms        map[string]*Room
	TotalClients int
	Mutex        sync.Mutex
}

// NewRoomManager initializes the manager and creates the default "Main" room.
func NewRoomManager() *RoomManager {
	rm := &RoomManager{
		Rooms: make(map[string]*Room),
	}
	mainRoom := NewRoom("Main")
	rm.Rooms["Main"] = mainRoom
	go Broadcaster(mainRoom.Broadcast, mainRoom.Clients, &mainRoom.Mutex)
	return rm
}

// RemoveClient fully disconnects a client and cleans up empty rooms.
func (rm *RoomManager) RemoveClient(currRoom *Room, conn net.Conn, name string) {
	if currRoom == nil {
		return
	}
	currRoom.Mutex.Lock()
	if clientStruct, ok := currRoom.Clients[conn]; ok && clientStruct.channel != nil {
		close(clientStruct.channel)
	}
	delete(currRoom.Clients, conn)
	currRoom.Mutex.Unlock()

	if name != "" {
		leaveMsg := fmt.Sprintf("User [%s] has left our chat...", name)
		currRoom.Mutex.Lock()
		currRoom.History = append(currRoom.History, leaveMsg)
		currRoom.Mutex.Unlock()
		currRoom.Broadcast <- Message{Text: leaveMsg, Sender: conn}
		log.Println("EVENT:", leaveMsg)
	}

	rm.Mutex.Lock()
	currRoom.Mutex.Lock()
	if len(currRoom.Clients) == 0 && currRoom.Name != "Main" {
		delete(rm.Rooms, currRoom.Name)
	}
	currRoom.Mutex.Unlock()
	rm.Mutex.Unlock()

	rm.BroadcastRoomsUpdate()
}

// CreateRoom creates a new room safely.
func (rm *RoomManager) CreateRoom(roomName string) error {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()
	if _, exists := rm.Rooms[roomName]; exists {
		return errors.New("Room already exists. Try joining it with /join=" + roomName)
	}
	newRoom := NewRoom(roomName)
	rm.Rooms[roomName] = newRoom
	go Broadcaster(newRoom.Broadcast, newRoom.Clients, &newRoom.Mutex)
	go rm.BroadcastRoomsUpdate()
	return nil
}

// ChangeRoom changes the room for a given connection and user safely.
func (rm *RoomManager) ChangeRoom(currRoom *Room, roomName string, conn net.Conn, name string) (*Room, error) {
	rm.Mutex.Lock()
	targetRoom, exists := rm.Rooms[roomName]
	rm.Mutex.Unlock()

	if !exists {
		return currRoom, errors.New("Room does not exist. Create it with /create=" + roomName)
	}

	if targetRoom == currRoom {
		return currRoom, errors.New("You are already in this room.")
	}

	targetRoom.Mutex.Lock()
	nameTaken := false
	for _, c := range targetRoom.Clients {
		if c.Name == name {
			nameTaken = true
			break
		}
	}
	if nameTaken {
		targetRoom.Mutex.Unlock()
		return currRoom, errors.New("Your name is already taken in that room. Change your name first with /name=")
	}
	targetRoom.Mutex.Unlock()

	leaveMsg := fmt.Sprintf("User [%s] has left our chat...", name)
	currRoom.Mutex.Lock()
	clientInfo := currRoom.Clients[conn]
	delete(currRoom.Clients, conn)
	currRoom.History = append(currRoom.History, leaveMsg)
	currRoom.Mutex.Unlock()
	currRoom.Broadcast <- Message{Text: leaveMsg, Sender: conn}

	rm.Mutex.Lock()
	currRoom.Mutex.Lock()
	if len(currRoom.Clients) == 0 && currRoom.Name != "Main" {
		delete(rm.Rooms, currRoom.Name)
	}
	currRoom.Mutex.Unlock()
	rm.Mutex.Unlock()

	targetRoom.Mutex.Lock()
	targetRoom.Clients[conn] = clientInfo
	targetRoom.Mutex.Unlock()

	rm.BroadcastRoomsUpdate()

	return targetRoom, nil
}

// ListRooms returns a formatted list of active rooms and their client counts.
func (rm *RoomManager) ListRooms() []string {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()
	var roomInfo []string
	for name, room := range rm.Rooms {
		room.Mutex.Lock()
		count := len(room.Clients)
		room.Mutex.Unlock()
		roomInfo = append(roomInfo, fmt.Sprintf("- %s (%d users)", name, count))
	}
	return roomInfo
}

// BroadcastRoomsUpdate sends a fresh rooms snapshot to all connected clients.
func (rm *RoomManager) BroadcastRoomsUpdate() {
	rooms := rm.ListRooms()

	rm.Mutex.Lock()
	recipients := make([]net.Conn, 0)
	for _, room := range rm.Rooms {
		room.Mutex.Lock()
		for conn := range room.Clients {
			recipients = append(recipients, conn)
		}
		room.Mutex.Unlock()
	}
	rm.Mutex.Unlock()

	for _, conn := range recipients {
		if _, err := fmt.Fprintln(conn, "Available rooms:"); err != nil {
			continue
		}
		for _, r := range rooms {
			if _, err := fmt.Fprintln(conn, r); err != nil {
				break
			}
		}
	}
}
