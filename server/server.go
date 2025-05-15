package server

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	port       int
	clients    map[*Client]bool
	rooms      map[string]*Room
	users      map[string]*User
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

type Message struct {
	Sender   string
	RoomName string
	Content  string
	Type     string // "text", "file", "command", "file-request", "file-chunk"
	FileData []byte // Used for file transfer
	FileName string // Used for file transfer
}

func NewServer(port int) *Server {
	return &Server{
		port:       port,
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*Room),
		users:      make(map[string]*User),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (s *Server) Run() error {
	// Start TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	defer listener.Close()

	// Create a default room
	s.rooms["general"] = NewRoom("general")

	// Log server startup
	log.Printf("Server started on port %d", s.port)

	// Start handling messages in a goroutine
	go s.handleMessages()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		log.Printf("New connection from %s", conn.RemoteAddr().String())

		// Create a new client
		client := NewClient(conn, s)

		// Handle the client in a separate goroutine
		go client.Handle()
	}
}

func (s *Server) handleMessages() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client] = true
			s.mutex.Unlock()
		case client := <-s.unregister:
			s.mutex.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mutex.Unlock()
		case message := <-s.broadcast:
			s.mutex.Lock()
			// If it's a room message, send only to clients in that room
			if message.RoomName != "" {
				room, exists := s.rooms[message.RoomName]
				if exists {
					room.Broadcast(message)
				}
			} else {
				// Otherwise, broadcast to all clients
				for client := range s.clients {
					if client.authenticated {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(s.clients, client)
						}
					}
				}
			}
			s.mutex.Unlock()
		}
	}
}

func (s *Server) RegisterUser(username, password string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.users[username]; exists {
		return false // User already exists
	}

	// Create and store the user
	s.users[username] = NewUser(username, password)
	return true
}

func (s *Server) AuthenticateUser(username, password string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, exists := s.users[username]
	if !exists {
		return false
	}

	return user.CheckPassword(password)
}
