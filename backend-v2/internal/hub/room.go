package hub

import (
	"fmt"
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
		Clients: map[string]*Client{},
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
