package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var messages = Messages{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub *Hub
	// Client details in Message type
	clientDetails Message
	// The websocket connection
	conn *websocket.Conn
	// JSON payload
	send chan Message
}

func (c *Client) readPump() {
	defer func() {
		for i, message := range c.hub.messages {
			if message.ID == c.clientDetails.ID {
				if message.Type == "ADD_USER" || message.Type == "UPDATE_USER" {
					// message.Type = "DELETE_USER"
					// fmt.Println("message", message)
					// SendJSON(c, message)
					// message = Message{}
					// messages[i] = messages[len(messages)-1]
					// messages = messages[:len(messages)-1]
					delete(c.hub.messages, i)
				}
			}
		}

		c.hub.unregister <- c
		c.conn.Close()
	}()
	// c.conn.SetReadLimit(maxMessageSize)
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			fmt.Println("error connection with client has been closed:", err)
			for i, message := range c.hub.messages {
				if message.ID == c.clientDetails.ID {
					if message.Type == "ADD_USER" || message.Type == "UPDATE_USER" {
						delete(c.hub.messages, i)
					}
				}
			}

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Attach ID to message to identify to whom it came from
		message.ID = c.clientDetails.ID

		switch message.Type {
		case "UPDATE_USER":
			c.clientDetails.Username = message.Username
			c.clientDetails.Type = message.Type
			c.clientDetails.Email = message.Email
			c.clientDetails.Time = message.Time
			c.hub.auth <- c
		default:
			c.hub.broadcast <- message
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		// ticker.Stop()
		c.conn.Close()
	}()
	for {
		message := <-c.send
		switch message.Type {
		case "ADD_MESSAGE":
			if c.clientDetails.ID == message.To {
				c.hub.messages[len(c.hub.messages)+1] = message
				// c.hub.messages = append(c.hub.messages, message)
				SendJSON(c, message)
			}
		case "MESSAGE_TO_ALL":
			c.hub.messages[len(c.hub.messages)+1] = message
			// c.hub.messages = append(c.hub.messages, message)
			SendJSON(c, message)
		case "ADD_USER":
			c.hub.messages[len(c.hub.messages)+1] = message
			// c.hub.messages = append(c.hub.messages, message)
			SendJSON(c, message)
		case "UPDATE_USER":
			c.hub.messages[len(c.hub.messages)+1] = message
			// c.hub.messages = append(c.hub.messages, message)
			SendJSON(c, message)
		}
	}
}

func SendJSON(c *Client, message Message) {
	err := c.conn.WriteJSON(message)
	if err != nil {
		log.Printf("failed to write JSON; Possible that ws connection has already closed: %v", err)
		// double check if already closing in hub
		c.conn.Close()
	}
}

func (c *Client) sendCachedMessages() {
	for _, message := range c.hub.messages {
		fmt.Println("range", message)
		switch message.Type {
		case "ADD_MESSAGE":
			if c.clientDetails.ID == message.To {
				SendJSON(c, message)
			}
		case "MESSAGE_TO_ALL":
			SendJSON(c, message)
		case "UPDATE_USER":
			fmt.Println("update_user", message)
			SendJSON(c, message)
		case "ADD_USER":
			fmt.Println("add_user", message)
			SendJSON(c, message)
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub: hub,
		clientDetails: Message{
			Username: "Anonymous",
			ID:       len(hub.clients) + 1,
			Email:    "anonymous@shoresuite.com",
			Type:     "ADD_USER",
			Time:     "",
		},
		conn: conn,
		send: make(chan Message),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writePump()
	go client.readPump()

	// Broadcast all messages in cache
	client.sendCachedMessages()
}
