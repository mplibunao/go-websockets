package main

type Hub struct {
	// Map of Registered Clients
	clients map[*Client]bool
	// Channel for inbound messages from clients
	broadcast chan Message
	// Channel for clients connecting to ws
	register chan *Client
	// Channel for clients disconnecting
	unregister chan *Client
	// Channel for clients logging in/out
	auth chan *Client
	// Consolidated list of messages since server start
	messages map[int]Message
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		auth:       make(chan *Client),
		messages:   make(map[int]Message),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			//client.send <- client.clientDetails
			for c := range h.clients {
				select {
				case c.send <- client.clientDetails:
				default:
					close(c.send)
					delete(h.clients, client)
				}
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case client := <-h.auth:
			// for c := range h.clients {
			// 	select {
			// 	case c.send <- client.clientDetails:
			// 	default:
			// 		close(c.send)
			// 		delete(h.clients, client)
			// 	}
			// }
			client.send <- client.clientDetails
		case message := <-h.broadcast:
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
