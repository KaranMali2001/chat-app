package hub

import (
	"fmt"
	"github.com/chat-app/pkg/logger"
	"sync"
)

type Room struct {
	Mutex   sync.RWMutex
	RoomId  string             `json:"room_id,omitempty"`
	Clients map[string]*Client `json:"clients,omitempty"`
}

func createRoom(roomId string) *Room {
	return &Room{
		RoomId:  roomId,
		Clients: make(map[string]*Client),
	}
}
func (r *Room) addClients(client *Client) error {

	if _, exist := r.Clients[client.Username]; exist {
		return fmt.Errorf("Client already exist in this")
	}
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.Clients[client.Username] = client
	return nil
}
func (r *Room) removeClient(c *Client) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if _, exist := r.Clients[c.Username]; !exist {
		return fmt.Errorf("Client %s Not Found in Room %s", c.Username, r.RoomId)
	}
	delete(r.Clients, c.Username)
	return nil
}
func (r *Room) Broadcast(event Event, exclude *Client) {
	r.Mutex.RLock()
	logger.Infof("Broadcasting to room with %d clients", len(r.Clients))
	clients := make([]*Client, 0, len(r.Clients))

	for _, c := range r.Clients {
		if exclude == nil || exclude.Username != c.Username {
			clients = append(clients, c)
			logger.Infof("Adding client %s to broadcast list", c.Username)
		}
	}
	r.Mutex.RUnlock()

	logger.Infof("Broadcasting event to %d clients", len(clients))
	for _, c := range clients {
		c.SendEvent(event)
	}
}
func (r *Room) getClientCount() int {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	return len(r.Clients)
}
func (h *Hub) GetRoomStats() map[string]interface{} {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_rooms"] = len(h.Rooms)
	stats["server_id"] = h.serverName

	roomStats := make(map[string]int)
	for roomID, room := range h.Rooms {
		roomStats[roomID] = room.getClientCount()
	}
	stats["room_clients"] = roomStats

	return stats
}
