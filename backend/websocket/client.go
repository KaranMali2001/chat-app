package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool
type Client struct {
	username  string
	conn      *websocket.Conn
	manager   *Manager
	egress    chan Event // used to avoid concurrency issue
	closeOnce sync.Once  //used to avoiding closing channel multiple times
}

var (
	pongWait     = time.Second * 10
	pingInterval = (pongWait * 9) / 10
)

func NewClient(conn *websocket.Conn, manager *Manager, username string) *Client {
	return &Client{
		conn:     conn,
		manager:  manager,
		egress:   make(chan Event, 256),
		username: username,
	}

}
func (c *Client) readMessage() {
	defer c.closeClient()
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.manager.log.Errorln("Error in setReadDeadline", err)
		return
	}
	c.conn.SetPongHandler(c.pongHandler)
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
	c.manager.log.Infoln("inside Write message")
	// defer c.closeClient()
	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case msg, ok := <-c.egress:
			if !ok {
				// if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
				// 	c.manager.log.Errorln("Error while sending close message to client", err)
				// 	return
				// }
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				c.manager.log.Errorln("Error while marshing the data", err)
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				c.manager.log.Errorw("error while writing msg to client", err)
				return // Exit on write error
			}
		case <-ticker.C:

			if err := c.conn.WriteMessage(websocket.PingMessage, []byte("")); err != nil {
				c.manager.log.Errorln("error while sending Ping Message", err)
				return
			}
		}
	}

}

func (c *Client) closeClient() {
	c.closeOnce.Do(func() {
		c.conn.SetWriteDeadline(time.Now().Add(time.Second))
		// c.conn.WriteMessage(websocket.CloseMessage, []byte{})
		close(c.egress)
		c.manager.log.Infoln("client egress channel closed")
		err := c.conn.Close()
		if err != nil {
			c.manager.log.Errorw("error while closing the connection still removing from manager", err)
		}
		c.manager.rw.Lock()
		delete(c.manager.clients, c)
		c.manager.rw.Unlock()
		c.manager.log.Infoln("client removed", c.conn.RemoteAddr())
		c.manager.log.Infoln("total clients are", len(c.manager.clients))
	})

}
func (c *Client) pongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	c.manager.log.Infoln("pong", pongMsg)
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
