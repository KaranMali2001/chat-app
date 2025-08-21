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
	maxReconnectAttempts = 3
	reconnectDelay       = 2 * time.Second
)

type Client struct {
	Username    string          `json:"username,omitempty"`
	Egress      chan Event      `json:"egress,omitempty"`
	CloseOnce   sync.Once       `json:"close_once,omitempty"`
	Conn        *websocket.Conn `json:"conn,omitempty"`
	Hub         *Hub
	closed      bool
	mu          sync.RWMutex
	Ctx         context.Context
	cancel      context.CancelFunc
	reconnect   chan struct{}
	roomID      string
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

func (c *Client) ensureConnection() error {
	c.mu.RLock()
	if c.Conn != nil {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	// Attempt to reconnect
	return c.reconnectWithRetry()
}

func (c *Client) reconnectWithRetry() error {
	var lastErr error
	
	for i := 0; i < maxReconnectAttempts; i++ {
		if i > 0 {
			time.Sleep(reconnectDelay)
		}
		
		// Create new connection
		conn, _, err := websocket.DefaultDialer.Dial(c.Conn.RemoteAddr().String(), nil)
		if err != nil {
			lastErr = err
			logger.Errorln("Reconnection attempt %d failed: %v", i+1, err)
			continue
		}
		
		c.mu.Lock()
		c.Conn = conn
		c.closed = false
		c.mu.Unlock()
		
		// Rejoin room if needed
		if c.roomID != "" {
			joinEvent := Event{
				Type: JOIN_ROOM,
				Payload: Message{
					RoomId: c.roomID,
					Sender: c.Username,
				},
			}
			if err := c.Hub.ProcessEvent(joinEvent, c); err != nil {
				logger.Errorln("Failed to rejoin room: %v", err)
				continue
			}
		}
		
		return nil
	}
	
	return fmt.Errorf("failed to reconnect after %d attempts: %v", maxReconnectAttempts, lastErr)
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
		logger.Infof("Write Go routine is terminated for client %s", c.Username)
	}()
	
	for {
		select {
		case <-c.Ctx.Done():
			logger.Infof("Context Done called in Write Go routine")
			return

		case event, ok := <-c.Egress:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// Ensure we have a valid connection
			if err := c.ensureConnection(); err != nil {
				logger.Errorln("Cannot send message, connection lost: %v", err)
				return
			}
			
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteJSON(event); err != nil {
				logger.Errorln("Error writing message to client %s: %v", c.Username, err)
				// Attempt to reconnect on write error
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					if reconnectErr := c.reconnectWithRetry(); reconnectErr == nil {
						// Retry sending the message after reconnection
						select {
						case c.Egress <- event:
							continue
						default:
							logger.Infof("Dropping message for %s: egress channel full", c.Username)
						}
					}
				}
				return
			}
			
		case <-ticker.C:
			if err := c.ensureConnection(); err != nil {
				logger.Errorln("Cannot send ping, connection lost: %v", err)
				return
			}
			
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Errorln("Error sending ping to client %s: %v", c.Username, err)
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
	c.mu.RUnlock()
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
