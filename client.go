package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// var messages = Messages{}
// var users = []Message

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
		for i := range c.hub.users {
			if i == c.clientDetails.ID {
				delete(c.hub.clients, c)
				//delete(c.hub.users, c.clientDetails.ID)
				c.hub.users[c.clientDetails.ID] = Message{}
				close(c.send)
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
			// for i, message := range c.hub.messages {
			// 	if message.ID == c.clientDetails.ID {
			// 		if message.Type == "ADD_USER" || message.Type == "UPDATE_USER" {
			// 			delete(c.hub.messages, i)
			// 		}
			// 	}
			// }

			for i := range c.hub.users {
				if i == c.clientDetails.ID {
					delete(c.hub.clients, c)
					//delete(c.hub.users, c.clientDetails.ID)
					c.hub.users[c.clientDetails.ID] = Message{}
					//close(c.send)
				}
			}

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Attach ID to message to identify to whom it came from
		message.ID = c.clientDetails.ID
		c.hub.broadcast <- message
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
				SendJSON(c, message)
			}
		case "MESSAGE_TO_ALL":
			c.hub.messages[len(c.hub.messages)+1] = message
			SendJSON(c, message)
		case "ADD_USER":
			c.hub.users = append(c.hub.users, message)
			SendJSON(c, message)
		case "UPDATE_USER":
			c.hub.messages[len(c.hub.messages)+1] = message
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
		}
	}
}

func (c *Client) sendUsers() {
	for _, user := range c.hub.users {
		if c.clientDetails.ID != user.ID {
			fmt.Println("sending user", user)
			SendJSON(c, user)
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
	client.sendUsers()
}
