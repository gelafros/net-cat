package handlers

import "net"

type Server struct {
	clientList map[net.Conn]string
	broadcast  chan Message
}

func NewServer() *Server {
	return &Server{
		clientList: make(map[net.Conn]string),
		broadcast:  make(chan Message),
	}
}
