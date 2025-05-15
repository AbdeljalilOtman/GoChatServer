package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	conn          net.Conn
	server        *Server
	send          chan Message
	username      string
	currentRoom   string
	authenticated bool
	fileBuffer    *bytes.Buffer
	receivingFile bool
	fileSize      int64
	receivedSize  int64
	fileName      string
}

func NewClient(conn net.Conn, server *Server) *Client {
	return &Client{
		conn:          conn,
		server:        server,
		send:          make(chan Message, 10),
		currentRoom:   "general", // Default room
		authenticated: false,
		fileBuffer:    new(bytes.Buffer),
		receivingFile: false,
	}
}

func (c *Client) Handle() {
	// Register client
	c.server.register <- c

	// Start goroutines for reading and writing
	go c.readPump()
	go c.writePump()

	// Send welcome message
	c.send <- Message{
		Sender:  "Server",
		Content: "Welcome to the chat server! Please log in with /login username password",
		Type:    "text",
	}
}

func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	reader := bufio.NewReader(c.conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from client:", err)
			}
			break
		}

		message = strings.TrimSpace(message)

		// Handle special commands
		if strings.HasPrefix(message, "/") {
			c.handleCommand(message)
			continue
		}

		// If not authenticated, don't allow sending messages
		if !c.authenticated {
			c.send <- Message{
				Sender:  "Server",
				Content: "You must log in first with /login username password",
				Type:    "text",
			}
			continue
		}

		// Broadcast the message to the current room
		c.server.broadcast <- Message{
			Sender:   c.username,
			RoomName: c.currentRoom,
			Content:  message,
			Type:     "text",
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		// Convert message to JSON
		jsonMsg, err := json.Marshal(message)
		if err != nil {
			log.Println("Error marshaling message:", err)
			continue
		}

		// Add newline character for message separation
		jsonMsg = append(jsonMsg, '\n')

		fmt.Printf("Sending to client %s: %s\n", c.username, string(jsonMsg))

		// Send message to client
		_, err = c.conn.Write(jsonMsg)
		if err != nil {
			log.Println("Error sending message to client:", err)
			break
		}

		// Debug log for sent messages
		log.Printf("Message sent to %s: %s", c.username, message.Content)
	}
}

func (c *Client) handleCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	fmt.Printf("Processing command: %s from client %s\n", cmd, c.username)

	log.Printf("Client %s executing command: %s", c.conn.RemoteAddr().String(), cmd)

	switch parts[0] {
	case "/login":
		if len(parts) != 3 {
			c.send <- Message{Sender: "Server", Content: "Usage: /login username password", Type: "text"}
			return
		}
		username, password := parts[1], parts[2]

		// Try to authenticate
		if c.server.AuthenticateUser(username, password) {
			c.username = username
			c.authenticated = true
			c.send <- Message{Sender: "Server", Content: "Login successful!", Type: "text"}
			// Join the default room
			c.server.rooms["general"].AddClient(c)
			log.Printf("User %s logged in successfully", username)
		} else {
			// Try to register if authentication fails
			if c.server.RegisterUser(username, password) {
				c.username = username
				c.authenticated = true
				c.send <- Message{Sender: "Server", Content: "Registered and logged in!", Type: "text"}
				// Join the default room
				c.server.rooms["general"].AddClient(c)
				log.Printf("User %s registered and logged in", username)
			} else {
				c.send <- Message{Sender: "Server", Content: "Invalid credentials", Type: "text"}
				log.Printf("Failed login attempt for user %s", username)
			}
		}

	case "/join":
		if !c.authenticated {
			c.send <- Message{Sender: "Server", Content: "You must log in first", Type: "text"}
			return
		}

		if len(parts) != 2 {
			c.send <- Message{Sender: "Server", Content: "Usage: /join roomname", Type: "text"}
			return
		}

		roomName := parts[1]

		// Create room if it doesn't exist
		c.server.mutex.Lock()
		if _, exists := c.server.rooms[roomName]; !exists {
			c.server.rooms[roomName] = NewRoom(roomName)
			fmt.Printf("Created new room: %s\n", roomName)
		}
		c.server.mutex.Unlock()

		// Leave current room
		if c.currentRoom != "" {
			if room, exists := c.server.rooms[c.currentRoom]; exists {
				room.RemoveClient(c)
				fmt.Printf("Client %s left room %s\n", c.username, c.currentRoom)
			}
		}

		// Join new room
		c.currentRoom = roomName
		if room, exists := c.server.rooms[roomName]; exists {
			room.AddClient(c)
			fmt.Printf("Client %s joined room %s\n", c.username, roomName)
		}

		// Send confirmation directly to the client
		directMsg := Message{
			Sender:  "Server",
			Content: "You have joined room: " + roomName,
			Type:    "text",
		}

		// Ensure the message is actually sent
		select {
		case c.send <- directMsg:
			fmt.Printf("Sent room join confirmation to %s\n", c.username)
		default:
			fmt.Printf("Failed to send room join confirmation to %s\n", c.username)
		}

	case "/rooms":
		if !c.authenticated {
			c.send <- Message{Sender: "Server", Content: "You must log in first", Type: "text"}
			return
		}

		c.server.mutex.Lock()
		roomList := "Available rooms:\n"
		for name := range c.server.rooms {
			roomList += "- " + name + "\n"
		}
		c.server.mutex.Unlock()

		fmt.Printf("Sending room list to client %s\n", c.username)
		c.send <- Message{Sender: "Server", Content: roomList, Type: "text"}

	case "/users":
		if !c.authenticated {
			c.send <- Message{Sender: "Server", Content: "You must log in first", Type: "text"}
			return
		}

		if c.currentRoom == "" {
			c.send <- Message{Sender: "Server", Content: "You are not in any room", Type: "text"}
			return
		}

		room, exists := c.server.rooms[c.currentRoom]
		if !exists {
			c.send <- Message{Sender: "Server", Content: "Room not found", Type: "text"}
			return
		}

		userList := fmt.Sprintf("Users in room %s:\n", c.currentRoom)

		room.mutex.Lock()
		for client := range room.clients {
			userList += "- " + client.username + "\n"
		}
		room.mutex.Unlock()

		fmt.Printf("Sending user list to client %s\n", c.username)
		c.send <- Message{Sender: "Server", Content: userList, Type: "text"}

	case "/sendfile":
		if !c.authenticated {
			c.send <- Message{Sender: "Server", Content: "You must log in first", Type: "text"}
			return
		}

		if len(parts) < 4 {
			c.send <- Message{Sender: "Server", Content: "Usage: /sendfile username filename filesize", Type: "text"}
			return
		}

		targetUser := parts[1]
		fileName := parts[2]
		fileSize, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			c.send <- Message{Sender: "Server", Content: "Invalid file size", Type: "text"}
			return
		}

		// Find the target user
		c.server.mutex.Lock()
		var recipient *Client
		for client := range c.server.clients {
			if client.username == targetUser && client.authenticated {
				recipient = client
				break
			}
		}
		c.server.mutex.Unlock()

		if recipient == nil {
			c.send <- Message{Sender: "Server", Content: "User not found or not online", Type: "text"}
			return
		}

		// Notify recipient about incoming file
		recipient.send <- Message{
			Sender:   c.username,
			Content:  fmt.Sprintf("Incoming file: %s (%.2f KB). Type /accept or /reject", fileName, float64(fileSize)/1024),
			Type:     "file-request",
			FileName: fileName,
		}

		c.send <- Message{
			Sender:  "Server",
			Content: "File transfer request sent. Waiting for " + targetUser + " to accept...",
			Type:    "text",
		}

	case "/accept":
		if !c.authenticated || !c.receivingFile {
			c.send <- Message{Sender: "Server", Content: "No pending file transfer", Type: "text"}
			return
		}

		c.send <- Message{Sender: "Server", Content: "File transfer accepted", Type: "text"}
		// Further implementation would set up file reception

	case "/reject":
		if !c.authenticated || !c.receivingFile {
			c.send <- Message{Sender: "Server", Content: "No pending file transfer", Type: "text"}
			return
		}

		c.send <- Message{Sender: "Server", Content: "File transfer rejected", Type: "text"}
		c.receivingFile = false

	default:
		c.send <- Message{Sender: "Server", Content: "Unknown command: " + parts[0], Type: "text"}
	}
}
