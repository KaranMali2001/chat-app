package hub

import (
	"context"
	"fmt"

	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/chat-app/pkg/logger"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = time.Second * 10
	pongWait       = time.Second * 60
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	egressBuffer   = 256
)

type Client struct {
	Username  string          `json:"username,omitempty"`
	Egress    chan Event      `json:"egress,omitempty"`
	CloseOnce sync.Once       `json:"close_once,omitempty"`
	Conn      *websocket.Conn `json:"conn,omitempty"`
	Hub       *Hub
	closed    bool
	mu        sync.RWMutex
	Ctx       context.Context
	cancel    context.CancelFunc
}

func NewClient(username string, conn *websocket.Conn, hub *Hub) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		Username: username,
		Egress:   make(chan Event, egressBuffer),
		Conn:     conn,
		Ctx:      ctx,
		Hub:      hub,
		cancel:   cancel,
	}
}
func (c *Client) ReadMessage() {
	defer func() {
		c.Close()
		logger.Infof(" Read Go Routine is terminated for client", c.Username)
	}()
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		select {
		case <-c.Ctx.Done():
			logger.Infof("Context Done called")
			return
		default:
			var event Event
			if err := c.Conn.ReadJSON(&event); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Errorln("Websocket client Error", c.Username, err)
				}

				logger.Errorln("Unknown Error")
				return
			}
			if err := c.Hub.ProcessEvent(event, c); err != nil {
				logger.Errorln("Error while processing event for client", c.Username, err)
				errorEvent := Event{
					Type: "error",
					Payload: Message{
						Id:      uuid.NewString(),
						Sender:  "SERVER",
						Content: fmt.Sprintf("Error: %v", err),
						Time:    time.Now().Format(time.RFC3339),
					},
				}
				select {
				case c.Egress <- errorEvent:
				default:
					logger.Infof("Client Egress Channel is Full")

				}
			}

		}
	}
}
func (c *Client) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
		logger.Infof("Write Go routine is terminated for client", c.Username)
	}()
	for {
		select {
		case <-c.Ctx.Done():
			logger.Info("Context Done called in Write Go routine")

		case event, ok := <-c.Egress:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteJSON(event); err != nil {
				logger.Errorln("Error while sending Message to Client", c.Username, err)
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Errorln("Error while sending Ping to client ", c.Username, err)
				return
			}
		}
	}

}
func (c *Client) Close() {
	c.CloseOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()

		c.cancel()
		close(c.Egress)
		c.Conn.Close()

		logger.Infof("Client %s closed", c.Username)
	})
}
func (c *Client) SendEvent(event Event) bool {
	c.mu.RLock()
	closed := c.closed
	if closed {
		logger.Infof("Client Egress channel is closed", closed)
		return false
	}
	select {
	case c.Egress <- event:
		return true
	default:
		logger.Logger.Sugar().Warnf("Client %s egress channle is full, Dropping Message", c.Username)
		return false
	}
}
