package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func handleClient(client net.Conn) {
	var message string

	client.Write([]byte("Welcome to TCP-Chat!\n"))
	client.Write([]byte("[ENTER YOUR NAME]: "))
	reader := bufio.NewReader(client)
	name, _ := reader.ReadString('n')
	name = strings.TrimSpace(name)
	message = strings.TrimSpace("Welcome to chat room, " + name + "\n")
	client.Write([]byte(message))

	// Message to Server with Print
	message = strings.TrimSpace("User \"" + name + "\" has enter the Chat Room!")
	fmt.Println(message)

	for {
		msg, err := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue // DO NOT broadcast empty messages
		}
		// Client Disconnected
		if err != nil {
			// fmt.Println("Error in reading Client input, Error:", err)
			fmt.Println(name, " has left our chat...")
			client.Close()
			break
		}
		// fmt.Println(client.RemoteAddr(), ":", msg)
		// fmt.Println(name, ":", msg)
		client.Write([]byte(msg))
	}
}

func main() {
	var port string

	args := os.Args[1:]
	if len(args) == 0 {
		port = ":8989"
	} else if len(args) == 1 {
		port = ":" + args[0]
	} else {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	// Create a server in "port". Listener is a handler for the server
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error in starting Server, Error: ", err)
		return
	}
	fmt.Println("Listening on the port", port)

	// Running loop for receiving multiple clients
	for {
		client, err := listener.Accept()
		if err != nil {
			fmt.Println("Error connecting Client, Error:", err)
			continue // for not terminating the Server
		}
		go handleClient(client)
		fmt.Println("A new client has joined:", client.RemoteAddr())
	}
}
