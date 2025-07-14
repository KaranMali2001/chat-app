package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Username  string          `json:"username,omitempty"`
	Egress    chan []Event    `json:"egress,omitempty"`
	CloseOnce sync.Once       `json:"close_once,omitempty"`
	Conn      *websocket.Conn `json:"conn,omitempty"`
}

func NewClient(username string, conn *websocket.Conn) *Client {
	return &Client{
		Username: username,
		Conn:     conn,
	}
}
