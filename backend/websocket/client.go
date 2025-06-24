package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool
type Client struct {
	conn      *websocket.Conn
	manager   *Manager
	egress    chan Event // used to avoid concurrency issue
	closeOnce sync.Once  //used to avoiding closing channel multiple times
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		egress:  make(chan Event),
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
		c.manager.log.Infoln("message type", msgType, " payload", string(payload))
		// go func() {
		// 	c.conn.WriteMessage(websocket.TextMessage, []byte("concurrent write!"))
		// }() //-> causes concurrent write problem
		// go c.broadcastMessage(payload)
		var req Event
		if err := json.Unmarshal(payload, &req); err != nil {
			c.manager.log.Desugar().Error("Error while unmarshing the json data")
		}

		if err := c.manager.routeEvent(req, c); err != nil {
			c.manager.log.Errorln("Error while routing the events", err)
		}

		// go func() {
		// 	c.conn.WriteMessage(websocket.TextMessage, []byte("concurrent write!"))
		// }() //-> does not cause concurrent write problem
	}
}

func (c *Client) writeMessage() {
	defer c.closeClient()
	for msg := range c.egress {
		c.manager.log.Infoln("Actual Event", msg)
		data, err := json.Marshal(msg)
		if err != nil {
			c.manager.log.Errorln("Error while marshing the data", err)
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			c.manager.log.Errorw("error while writing msg to client", err)
			return // Exit on write error
		}
		c.manager.log.Infoln("Message sent")
	}

}

// func (c *Client) broadcastMessage(payload []byte) {
// 	c.manager.rw.Lock()
// 	clients := make([]*Client, 0, len(c.manager.clients))
// 	for client := range c.manager.clients {
// 		if client == c {
// 			c.manager.log.Infoln("Skipping Current Client", c.conn.RemoteAddr())
// 			continue
// 		}
// 		clients = append(clients, client)
// 	}
// 	c.manager.rw.Unlock()
// 	finalPayload := string(payload) + " from server"
// 	c.manager.log.Info("total client to be brodcasted ", len(clients))
// 	for _, client := range clients {
// 		select {
// 		case client.egress <- []byte(finalPayload):
// 			c.manager.log.Infoln("Message added to engress")
// 		default:
// 			c.manager.log.Warnln("Client is slow and egress is full for client", client.conn.RemoteAddr())
// 		}
// 	}

// }
func (c *Client) closeClient() {
	c.closeOnce.Do(func() {
		c.conn.SetWriteDeadline(time.Now().Add(time.Second))
		c.conn.WriteMessage(websocket.CloseMessage, []byte{})
		close(c.egress)
		c.manager.log.Infoln("client egress channel closed")
		err := c.conn.Close()
		if err != nil {
			c.manager.log.Errorw("error while closing the connection still removing from manager", err)
		}
		c.manager.removeClient(c)
		c.manager.log.Infoln("client removed", c.conn.RemoteAddr())
		c.manager.log.Infoln("total clients are", len(c.manager.clients))
	})

}
