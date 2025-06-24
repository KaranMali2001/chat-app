package websocket

import (
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("origin")

			return origin == os.Getenv("ALLOWED_ORIGINS")
		},
	}
)

type Manager struct {
	clients  ClientList
	rw       sync.RWMutex
	log      *zap.SugaredLogger
	handlers map[string]EventHandler
}

func NewManager(logger *zap.SugaredLogger) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		log:      logger,
		handlers: make(map[string]EventHandler),
	}
	m.setUpEventHandlers()
	return m
}
func (m *Manager) setUpEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}
func SendMessage(event Event, c *Client) error {
	event.Payload.Content = "SEVER :" + event.Payload.Content
	event.Payload.Sender = "SERVER"

	event.Payload.Time = time.Now().Format("03:04 PM")
	c.egress <- event
	return nil

}
func (m *Manager) routeEvent(event Event, c *Client) error {
	if hanlder, ok := m.handlers[event.Type]; ok {
		if err := hanlder(event, c); err != nil {
			m.log.Errorln("Error while executing the handler", err)
			return err
		}
		return nil
	} else {
		m.log.Warnln("Event Hanlder for event ", event.Type, " Not found")
		return errors.New("no hanlder found")
	}
}

func (m *Manager) ServeWs(w http.ResponseWriter, r *http.Request) {

	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.log.Errorw("error while upgrading the websocket", err)
	}
	client := NewClient(conn, m)
	m.addClient(client)
	m.log.Infoln("New client connected ")
	m.log.Infoln("total clients are", len(m.clients))
	go client.readMessage()

	go client.writeMessage()

}
func (m *Manager) addClient(c *Client) error {
	m.rw.Lock()
	defer m.rw.Unlock()
	//will be addding more logic to add client to db or redis not just in memory object
	m.clients[c] = true
	return nil
}
func (m *Manager) removeClient(c *Client) error {
	m.rw.Lock()
	defer m.rw.Unlock()
	if _, ok := m.clients[c]; ok {
		c.conn.Close()
		delete(m.clients, c)
	}
	return nil
}
