package websocket

import ()

type Event struct {
	Type    string  `json:"type,omitempty"`
	Payload Message `json:"payload,omitempty"`
}
type Message struct {
	Id      string `json:"id"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Time    string `json:"time"`
}
type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage = "SEND_MESSAGE"
)

type SendMessageEvent struct {
	Message string `json:"message,omitempty"`
	From    string `json:"from,omitempty"`
}
