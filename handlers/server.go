package handlers

import "net"

type Server struct {
	clientList map[net.Conn]string
	broadcast  chan string
}

func NewServer() *Server {
	return &Server{
		clientList: make(map[net.Conn]string),
		broadcast:  make(chan string),
	}
}
