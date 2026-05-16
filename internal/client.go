package internal

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Message struct to hold the message text and identify the sender
type Message struct {
	Text   string
	Sender net.Conn
}

// Client struct to hold the users name and their dedicated message channel
type Client struct {
	Name    string
	channel chan string
}

// Struct to pass the updated state to the clientWriter safely
type clientState struct {
	colorCode string
	name      string
}

// Background writer goroutine dedicated to a single client
func clientWriter(conn net.Conn, ch chan string, stateChan chan clientState, initialState clientState) {
	state := initialState
	for {
		select {
		case newState := <-stateChan:
			state = newState
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if !strings.HasSuffix(msg, "\n") {
				msg += "\n"
			}
			// Clear the current line, print the message, and actively redraw the prompt
			prompt := fmt.Sprintf("%s[%s][%s]:\033[0m", state.colorCode, time.Now().Format("2006-01-02 15:04:05"), state.name)
			fmt.Fprintf(conn, "\r\033[2K%s%s", msg, prompt)
		}
	}
}

// Handles conns,names,chat history,and broadcasts names within specific rooms
func HandleClient(conn net.Conn, manager *RoomManager) {
	var name string

	// Default color is standard terminal color (ANSI reset code)
	colorCode := "\033[0m"

	// Capture the client's network address for our server logs
	addr := conn.RemoteAddr().String()

	manager.Mutex.Lock()
	currRoom := manager.Rooms["Main"]
	manager.Mutex.Unlock()

	defer func() {
		manager.Mutex.Lock()
		manager.TotalClients--
		manager.Mutex.Unlock()

		manager.RemoveClient(currRoom, conn, name)
		conn.Close()
	}()

	// Welcome message and ASCII art logo (Original NetCat Linux "penguin")
	logo := "         _nnnn_\n" +
		"        dGGGGMMb\n" +
		"       @p~qp~~qMb\n" +
		"       M|@||@) M|\n" +
		"       @,----.JM|\n" +
		"      JS^\\__/  qKL\n" +
		"     dZP        qKRb\n" +
		"    dZP          qKKb\n" +
		"   fZP            SMMb\n" +
		"   HZM            MMMM\n" +
		"   FqM            MMMM\n" +
		" __| \".        |\\dS\"qML\n" +
		" |    `.       | `' \\Zq\n" +
		"_)      \\.___.,|     .'\n" +
		"\\____   )MMMMMP|   .'\n" +
		"     `-'       `--'\n" +
		"[ENTER YOUR NAME]: "
	if _, err := fmt.Fprint(conn, logo); err != nil {
		log.Printf("Write error to %s: %v\n", addr, err)
		return
	}

	// Scanner to easily read text line-by-line from the clients conn
	msgscanner := bufio.NewScanner(conn)
	// Exit immediately if the client disconnects before entering anything
	if !msgscanner.Scan() {
		return
	}

	var stateChan chan clientState

	// The username chosen by the client
	name = msgscanner.Text()
	// Keep looping until a valid and unique name is provided
	for {
		if name == "" {
			if _, err := fmt.Fprint(conn, "\nInvalid name, try again\n[ENTER YOUR NAME]: "); err != nil {
				log.Printf("Write error to %s: %v\n", addr, err)
				return
			}
		} else {
			currRoom.Mutex.Lock()
			isTaken := false
			for _, existingClient := range currRoom.Clients {
				if existingClient.Name == name {
					isTaken = true
					break
				}
			}
			if !isTaken {
				clientChannel := make(chan string, 100) // Give them a mailbox that holds up to 100 unread messages
				currRoom.Clients[conn] = Client{Name: name, channel: clientChannel}
				currRoom.Mutex.Unlock()

				stateChan = make(chan clientState, 10)
				// Start the background writer specifically for this client
				go clientWriter(conn, clientChannel, stateChan, clientState{colorCode: colorCode, name: name})
				break
			}
			currRoom.Mutex.Unlock()
			if _, err := fmt.Fprint(conn, "\nName already taken, try another\n[ENTER YOUR NAME]: "); err != nil {
				log.Printf("Write error to %s: %v\n", addr, err)
				return
			}
		}
		if !msgscanner.Scan() {
			return
		}
		name = msgscanner.Text()
	}

	// Print the epic Zone01 and Devs ASCII art now that they've successfully joined
	welcomeHeader := fmt.Sprintf("\033[1;35m"+`
       __      __ ___  _     ___  ___  __  __  ___ 
       \ \    / /| __|| |   / __|/ _ \|  \/  || __|
        \ \/\/ / | _| | |__| (__| (_) | |\/| || _| 
         \_/\_/  |___||____|\___|\___/|_|  |_||___|
                                             
                 TO ZONE01 NETCAT SERVER
                       >>> %s <<<
`+"\033[0m", name)

	zone01Logo := welcomeHeader + "\033[1;36m" + `
 ███████╗ ██████╗ ███╗   ██╗███████╗ ██████╗  ██╗
 ╚══███╔╝██╔═══██╗████╗  ██║██╔════╝██╔═████╗███║
   ███╔╝ ██║   ██║██╔██╗ ██║█████╗  ██║██╔██║╚██║
  ███╔╝  ██║   ██║██║╚██╗██║██╔══╝  ████╔╝██║ ██║
 ███████╗╚██████╔╝██║ ╚████║███████╗╚██████╔╝ ██║
 ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝ ╚═════╝  ╚═╝

       _ _   ___  ____  ___   __   ____   /\_/\
      ( \( )(  _)(_  _)/ __) / _\ (_  _) ( o.o )
       ) (   ) _)  )( ( (__ /    \  )(    > ^ <
      (_)\_)(___) (__) \___)\_/\_/ (__)

                      CREATED BY:

         _  _ _____   _    _     _      _   ___ ___ ___ 
        | |/ /_   _| /_\  | |   | |    /_\ | _ \_ _/ __|
        | ' <  | |  / _ \ | |__ | |__ / _ \|   /| |\__ \
        |_|\_\ |_| /_/ \_\|____||____/_/ \_\|_|_\___|___/

        ___ ___ _      _   ___ ___  ___  ___ 
       / __| __| |    /_\ | __| _ \/ _ \/ __|
      | (_ | _|| |__ / _ \| _||   / (_) \__ \
       \___|___|____/_/ \_\_| |_|_\\___/|___/

         ___ ___ ___ ___ ___  ___   _   ___ 
        |   \_ _| _ \_ _/ __| /_\ / __|
        | |) | ||   /| || (_ |/ _ \\__ \
        |___/___|_|_\___|\___/_/ \_\___/
` + "\033[0m"
	fmt.Fprintln(conn, zone01Logo)

	// Print the past chat history for the new user
	currRoom.Mutex.Lock()
	for _, historyMsg := range currRoom.History {
		if _, err := fmt.Fprintln(conn, historyMsg); err != nil {
			log.Printf("Write error to %s: %v\n", addr, err)
			break
		}
	}
	currRoom.Mutex.Unlock()

	// Let the user know how to see available commands
	if _, err := fmt.Fprint(conn, "Type /help for instructions\n\n"); err != nil {
		log.Printf("Write error to %s: %v\n", addr, err)
		return
	}

	// Announce the new client to the rest of the chat room
	joinMsg := fmt.Sprintf("User [%s] has joined our chat...", name)
	currRoom.Mutex.Lock()
	currRoom.History = append(currRoom.History, joinMsg)
	currRoom.Mutex.Unlock()
	log.Println("EVENT:", joinMsg) // Save join event to file

	// Send the join message to the central broadcast to send to receivers
	currRoom.Broadcast <- Message{Text: joinMsg, Sender: conn}

	// Announce the updated room capacities to all users now that the user has fully joined
	manager.BroadcastRoomsUpdate()

	// Wait for new messages from this client
	for msgscanner.Scan() {
		// The actual msg typed by the client
		msg := msgscanner.Text()

		// Robustly filter out all ANSI escape sequences and control characters
		var cleanInput strings.Builder
		inEscape := false
		for _, r := range msg {
			if r == '\033' { // Escape character detected
				inEscape = true
				continue
			}
			if inEscape {
				// ANSI sequences typically end with a letter or a tilde
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '~' {
					inEscape = false
				}
				continue
			}
			// Ignore backspace and delete characters so they don't break formatting
			if r == '\b' || r == '\x7f' {
				continue
			}
			cleanInput.WriteRune(r)
		}
		msg = cleanInput.String()

		// Move the cursor up and erase the raw input line from their terminal
		fmt.Fprint(conn, "\033[1A\r\033[2K")

		// Ignore empty lines to prevent chat spam
		if strings.TrimSpace(msg) == "" {
			// Redraw the prompt if they send an empty message
			fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user wants to see the help menu
		if isHelpCmd := HandleHelp(msg, conn); isHelpCmd {
			fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		if isRoomsCmd := handleRooms(msg, conn, manager); isRoomsCmd {
			continue
		}

		if isCreateCmd := handleCreate(msg, conn, manager); isCreateCmd {
			continue
		}

		isJoinCmd, newRoomJoin := handleJoin(msg, conn, name, currRoom, manager)
		if isJoinCmd {
			currRoom = newRoomJoin
			continue
		}

		isLeaveCmd, newRoomLeave := handleLeave(msg, conn, name, currRoom, manager)
		if isLeaveCmd {
			currRoom = newRoomLeave
			continue
		}

		// Check if the user is trying to change their color
		if isColorCmd := handleColor(msg, conn, &colorCode, name, &currRoom.Mutex, currRoom.Broadcast, &currRoom.History); isColorCmd {
			stateChan <- clientState{colorCode: colorCode, name: name}
			fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user is trying to change their name
		if isNameCmd := handleName(msg, conn, &name, currRoom.Clients, &currRoom.Mutex, currRoom.Broadcast, &currRoom.History); isNameCmd {
			stateChan <- clientState{colorCode: colorCode, name: name}
			fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		if isWhisper := handleWhisper(msg, conn, name, currRoom.Clients, &currRoom.Mutex); isWhisper {
			fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user wants to flip a table
		msg = HandleTableFlip(msg)

		// Format AND CLEAN the message and send it to the central broadcast channel.
		cleanMsg := fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), name, msg)
		formattedMsg := fmt.Sprintf("%s%s\033[0m", colorCode, cleanMsg)

		// Echo the fully formatted and colored message back to the sender!
		fmt.Fprintln(conn, formattedMsg)

		currRoom.Mutex.Lock()
		currRoom.History = append(currRoom.History, formattedMsg)
		currRoom.Mutex.Unlock()
		currRoom.Broadcast <- Message{Text: formattedMsg, Sender: conn}

		// Save to file without color codes AND without the duplicate timestamp
		log.Printf("CHAT: [%s]:%s\n", name, msg)

		// Redraw the prompt for the sender after they send a message
		fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
	}

	// Log if the scanner stopped due to a network error instead of EOF
	if err := msgscanner.Err(); err != nil {
		log.Printf("Read error from %s: %v\n", addr, err)
	}
}
