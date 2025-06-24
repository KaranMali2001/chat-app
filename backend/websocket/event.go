package websocket

type Event struct {
	Type    string  `json:"type"`
	Payload Message `json:"payload"`
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
