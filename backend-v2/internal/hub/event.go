package hub

import (
	"time"

	"github.com/google/uuid"
)

const (
	CREATE_ROOM      = "create_room"
	SEND_MESSAGE     = "send_message"
	MESSAGE_RECEVIED = "message_recevied"
	LEAVE_ROOM       = "leave_room"
	JOIN_ROOM        = "join_room"
	USER_JOINED      = "user_joined"
	USER_LEFT        = "user_left"
	ERROR            = "error"
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
type RedisMessage struct {
	Event    Event  `json:"event,omitempty"`
	RoomId   string `json:"room_id,omitempty"`
	ServerId string `json:"server_id,omitempty"`
}

func NewMessage(sender, content, roomId string) Message {
	return Message{
		Id:      uuid.NewString(),
		Sender:  sender,
		Content: content,
		RoomId:  roomId,
		Time:    time.Now().Format(time.RFC3339),
	}
}
