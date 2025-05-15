package server

import (
	"fmt"
	"sync"
)

type Room struct {
	name    string
	clients map[*Client]bool
	mutex   sync.Mutex
}

func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[*Client]bool),
	}
}

func (r *Room) AddClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.clients[client] = true

	// Broadcast to room that a new user has joined
	fmt.Printf("User %s joined room %s\n", client.username, r.name)

	msg := Message{
		Sender:   "Server",
		RoomName: r.name,
		Content:  client.username + " has joined the room",
		Type:     "text",
	}

	// Broadcast to all clients in the room
	for c := range r.clients {
		if c.authenticated && c != client {
			select {
			case c.send <- msg:
			default:
				// If the client's send channel is full, remove them
				delete(r.clients, c)
			}
		}
	}
}

func (r *Room) RemoveClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.clients[client]; ok {
		delete(r.clients, client)

		// Broadcast to room that a user has left
		fmt.Printf("User %s left room %s\n", client.username, r.name)

		msg := Message{
			Sender:   "Server",
			RoomName: r.name,
			Content:  client.username + " has left the room",
			Type:     "text",
		}

		// Broadcast to remaining clients in the room
		for c := range r.clients {
			if c.authenticated {
				select {
				case c.send <- msg:
				default:
					// If the client's send channel is full, remove them
					delete(r.clients, c)
				}
			}
		}
	}
}

func (r *Room) Broadcast(message Message) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for client := range r.clients {
		if client.authenticated {
			select {
			case client.send <- message:
			default:
				// If the client's send channel is full, remove them
				delete(r.clients, client)
			}
		}
	}
}
