package main

type Message struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Type     string `json:"type"`
	To       int    `json:"to"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

type Messages []Message

// func (m Messages) addMessage(msg Message) {
// 	m = append(m, msg)
// }

// func newMessages() *Messages {
// 	return &Messages{}
// }
