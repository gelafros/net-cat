package internal

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
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

// Background writer goroutine dedicated to a single client
func clientWriter(conn net.Conn, ch chan string, name *string, colorCode *string) {
	for msg := range ch {
		// Clear line, print the message, and reprint the prompt
		fmt.Fprintln(conn, "\r\033[2K"+msg)
		fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", *colorCode, time.Now().Format("2006-01-02 15:04:05"), *name)
	}
}

// Handles conns,names,chat history,and broadcasts names
func HandleClient(conn net.Conn, clients map[net.Conn]Client, key *sync.Mutex, broadcast chan Message, history *[]string) {
	var name string

	// Default color is standard terminal color (ANSI reset code)
	colorCode := "\033[0m"

	// Capture the client's network address for our server logs
	addr := conn.RemoteAddr().String()

	defer func() {
		// Waits until the whole func is run and safely removes the client from the clients map
		key.Lock()
		if clientStruct, ok := clients[conn]; ok && clientStruct.channel != nil {
			// Close the channel so the clientWriter goroutine stops cleanly
			close(clientStruct.channel)
		}
		delete(clients, conn)
		key.Unlock()

		// If the user successfully joined
		if name != "" {
			leaveMsg := fmt.Sprintf("User [%s] has left our chat...", name)
			key.Lock()
			*history = append(*history, leaveMsg)
			key.Unlock()
			broadcast <- Message{Text: leaveMsg, Sender: conn}
			log.Println("EVENT:", leaveMsg) // Save leave event to file
		}
		// log.Printf("Client disconnected: %s\n", addr)
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

	// The username chosen by the client
	name = msgscanner.Text()
	// Keep looping until a valid and unique name is provided
	for {
		if name == "" {
			if _, err := fmt.Fprintln(conn, "Invalid name, try again stupid"); err != nil {
				log.Printf("Write error to %s: %v\n", addr, err)
				return
			}
			fmt.Fprint(conn, "[ENTER YOUR NAME]: ")
		} else {
			// Safely check if the name is already used
			key.Lock()
			isTaken := false
			for _, existingClient := range clients {
				if existingClient.Name == name {
					isTaken = true
					break
				}
			}
			if !isTaken {
				// Initialize the struct and its dedicated channel
				clientChannel := make(chan string, 100) // Give them a mailbox that holds up to 100 unread messages
				clients[conn] = Client{Name: name, channel: clientChannel}

				key.Unlock()

				// Start the background writer specifically for this client
				go clientWriter(conn, clientChannel, &name, &colorCode)
				break
			}
			key.Unlock()
			if _, err := fmt.Fprintln(conn, "Name already taken, try another"); err != nil {
				log.Printf("Write error to %s: %v\n", addr, err)
				return
			}
			fmt.Fprint(conn, "[ENTER YOUR NAME]: ")
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

       _  _  ___  ____  ___   __   ____   /\_/\
      ( \( )/ __)(_  _)/ __) / _\ (_  _) ( o.o )
       )  ( \__ \  )( ( (__ /    \  )(    > ^ <
      (_)\_)(___/ (__) \___)\_/\_/ (__)

                      CREATED BY:

         _  _____  _   _    _   ___ ___ ___ 
        | |/ /_   _| \| |  /_\ | _ \_ _/ __|
        | ' <  | | | .  | / _ \|   /| |\__ \
        |_|\_\ |_| |_|\_|/_/ \_\_|_\___|___/

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

	// Let the user know how to see available commands
	if _, err := fmt.Fprintln(conn, "Type /help for instructions"); err != nil {
		log.Printf("Write error to %s: %v\n", addr, err)
		return
	}

	// Print the past chat history for the new user
	key.Lock()
	for _, historyMsg := range *history {
		if _, err := fmt.Fprintln(conn, historyMsg); err != nil {
			log.Printf("Write error to %s: %v\n", addr, err)
			break
		}
	}
	key.Unlock()

	// Announce the new client to the rest of the chat room
	joinMsg := fmt.Sprintf("User [%s] has joined our chat...", name)
	key.Lock()
	*history = append(*history, joinMsg)
	key.Unlock()
	log.Println("EVENT:", joinMsg) // Save join event to file

	// Send the join message to the central broadcast to send to receivers
	broadcast <- Message{Text: joinMsg, Sender: conn}

	// Print the initial typing prompt for the new user
	fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)

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

		// Check if the user is trying to change their color
		if isColorCmd := handleColor(msg, conn, &colorCode, name, key, broadcast, history); isColorCmd {
			fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		// Check if the user is trying to change their name
		if isNameCmd := handleName(msg, conn, &name, clients, key, broadcast, history); isNameCmd {
			fmt.Fprintf(conn, "%s[%s][%s]:\033[0m", colorCode, time.Now().Format("2006-01-02 15:04:05"), name)
			continue
		}

		if isWhisper := handleWhisper(msg, conn, name, clients, key); isWhisper {
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

		key.Lock()
		*history = append(*history, formattedMsg)
		key.Unlock()
		broadcast <- Message{Text: formattedMsg, Sender: conn}

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
