package main

import "fmt"

type Hub struct {
	// Map of Registered Clients
	clients map[*Client]bool
	// Channel for inbound messages from clients
	broadcast chan Message
	// Channel for clients connecting to ws
	register chan *Client
	// Channel for clients disconnecting
	unregister chan *Client
	// Consolidated list of messages since server start
	messages map[int]Message
	// List of online users
	users []Message
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		messages:   make(map[int]Message),
		users:      make([]Message, 0),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// client.send <- client.clientDetails
			// for c := range h.clients {
			// 	select {
			// 	case c.send <- client.clientDetails:
			// 	default:
			// 		close(c.send)
			// 		delete(h.clients, client)
			// 	}
			// }
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				h.users[client.clientDetails.ID] = Message{}
				close(client.send)
			}
		case message := <-h.broadcast:
			fmt.Println("message broadcast", message)
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
