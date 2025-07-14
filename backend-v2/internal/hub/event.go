package hub

const (
	CREATE_ROOM  = "create_room"
	SEND_MESSAGE = "send_message"
	LEAVE_ROOM   = "leave_room"
	JOIN_ROOM    = "join_room"
)

type Event struct {
	Type    string  `json:"type"`
	Payload Message `json:"payload"`
}

type Message struct {
	Id      string `json:"id"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
	RoomId  string `json:"room_id"` // Add this for room context
	Time    string `json:"time"`
}
type EventHandler func(event Event, client *Client) error
