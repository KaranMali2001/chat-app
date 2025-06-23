package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool
type Client struct {
	conn    *websocket.Conn
	manager *Manager
	egress  chan []byte // used to avoid concurrency issue
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		egress:  make(chan []byte, 256),
	}

}
func (c *Client) readMessage() {
	defer c.closeClient()
	for {
		msgType, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {

				c.manager.log.Errorw("error while reading message", err)
			}
			c.manager.log.Infoln("Noramlly closed ", err)
			break

		}
		// go func() {
		// 	c.conn.WriteMessage(websocket.TextMessage, []byte("concurrent write!"))
		// }() //-> causes concurrent write problem
		go c.broadcastMessage(payload)
		c.manager.log.Infoln("message type", msgType, " payload", string(payload))
		// go func() {
		// 	c.conn.WriteMessage(websocket.TextMessage, []byte("concurrent write!"))
		// }() //-> does not cause concurrent write problem
	}
}

func (c *Client) writeMessage() {
	defer c.closeClient()
	for msg := range c.egress {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			c.manager.log.Errorw("error while writing msg to client", err)
			return // Exit on write error
		}
		c.manager.log.Infoln("Message sent")
	}

	// Channel closed, send close message
	if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
		c.manager.log.Errorw("error in sending close message to client", err)
	}
}

func (c *Client) broadcastMessage(payload []byte) {
	c.manager.rw.Lock()
	clients := make([]*Client, 0, len(c.manager.clients))
	for client := range c.manager.clients {
		if client == c {
			c.manager.log.Infoln("Skipping Current Client", c.conn.RemoteAddr())
			continue
		}
		clients = append(clients, client)
	}
	c.manager.rw.Unlock()
	finalPayload := string(payload) + " from server"
	c.manager.log.Info("total client to be brodcasted ", len(clients))
	for _, client := range clients {
		select {
		case client.egress <- []byte(finalPayload):
			c.manager.log.Infoln("Message added to engress")
		default:
			c.manager.log.Warnln("Client is slow and egress is full for client", client.conn.RemoteAddr())
		}
	}

}
func (c *Client) closeClient() {
	c.conn.SetWriteDeadline(time.Now().Add(time.Second))
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	close(c.egress)
	err := c.conn.Close()
	if err != nil {
		c.manager.log.Errorw("error while closing the connection still removing from manager", err)
	}
	c.manager.removeClient(c)
	c.manager.log.Infoln("client removed", c)
	c.manager.log.Infoln("total clients are", len(c.manager.clients))
}
