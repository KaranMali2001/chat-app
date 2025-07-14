package hub

import (
	"fmt"
	"sync"
)

type Hub struct {
	Mu       sync.RWMutex
	Rooms    map[string]*Room
	Handlers map[string]EventHandler
}

func (h *Hub) RegisterHandlers(event_type string, handler EventHandler) {
	h.Handlers[event_type] = handler
}
func (h *Hub) ProcessEvent(event Event, client *Client) error {
	handler, exist := h.Handlers[event.Type]
	if !exist {
		return fmt.Errorf("handler not found")
	}
	return handler(event, client)
}
func (h *Hub) RegisterDefaultHandlers() {

	h.RegisterHandlers(SEND_MESSAGE, h.HandleSendMessage)
	h.RegisterHandlers(CREATE_ROOM, h.HandleCreateRoom)
	h.RegisterHandlers(LEAVE_ROOM, h.HandleLeaveRoom)
	h.RegisterHandlers(JOIN_ROOM, h.HandleJoinRoom)

}
func (h *Hub) HandleCreateRoom(event Event, client *Client) error {
	roomId := event.Payload.RoomId
	if roomId == "" {
		return fmt.Errorf("missing room id")
	}
	h.Mu.Lock()
	defer h.Mu.Unlock()
	if _, exist := h.Rooms[roomId]; exist {
		return fmt.Errorf("room already exist ,try to join the room")
	}
	newRoom := createRoom(roomId)

	h.Rooms[roomId] = newRoom
	return nil
}

func (h *Hub) HandleSendMessage(event Event, client *Client) error {
	return nil
}
func (h *Hub) HandleLeaveRoom(event Event, client *Client) error {
	return nil
}
func (h *Hub) HandleJoinRoom(event Event, client *Client) error {
	h.Mu.RLock()
	roomId := event.Payload.RoomId
	if _, exist := h.Rooms[roomId]; !exist {
		return fmt.Errorf("room does not exist")
	}
	room := h.Rooms[roomId]
	room.addClients(client)
	return nil
}
func NewHub() *Hub {

	h := &Hub{
		Rooms:    make(map[string]*Room),
		Handlers: make(map[string]EventHandler),
	}
	h.RegisterDefaultHandlers()
	return h
}
