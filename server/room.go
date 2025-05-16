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
	fmt.Printf("Added %s to room %s\n", client.username, r.name)

	// Broadcast to room that a new user has joined, but not to the new user
	for c := range r.clients {
		if c != client && c.authenticated {
			c.directSend(Message{
				Sender:   "Server",
				RoomName: r.name,
				Content:  client.username + " has joined the room",
				Type:     "text",
			})
		}
	}
}

func (r *Room) RemoveClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.clients[client]; ok {
		delete(r.clients, client)
		fmt.Printf("Removed %s from room %s\n", client.username, r.name)

		// Broadcast to room that a user has left
		for c := range r.clients {
			if c.authenticated {
				c.directSend(Message{
					Sender:   "Server",
					RoomName: r.name,
					Content:  client.username + " has left the room",
					Type:     "text",
				})
			}
		}
	}
}

func (r *Room) Broadcast(message Message) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	fmt.Printf("Broadcasting in room %s: %s\n", r.name, message.Content)

	for client := range r.clients {
		if client.authenticated {
			client.directSend(message)
		}
	}
}
