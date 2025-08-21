package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/chat-app/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	Mu          sync.RWMutex
	Rooms       map[string]*Room
	Handlers    map[string]EventHandler
	redisClient *redis.Client
	serverName  string
	pubsub      *redis.PubSub
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewHub(redisClient *redis.Client, serverName string) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Hub{
		Rooms:       make(map[string]*Room),
		Handlers:    make(map[string]EventHandler),
		redisClient: redisClient,
		serverName:  serverName,
		ctx:         ctx,
		cancel:      cancel,
	}
	h.RegisterDefaultHandlers()
	h.startRedisSubscriber()
	return h
}

func (h *Hub) RegisterHandlers(event_type string, handler EventHandler) {
	h.Handlers[event_type] = handler
}
func (h *Hub) ProcessEvent(event Event, client *Client) error {
	handler, exist := h.Handlers[event.Type]
	if !exist {
		return fmt.Errorf("handler not found for event Type %s", event.Type)
	}
	return handler(event, client)
}
func (h *Hub) RegisterDefaultHandlers() {

	h.RegisterHandlers(SEND_MESSAGE, h.HandleSendMessage)
	h.RegisterHandlers(CREATE_ROOM, h.HandleCreateRoom)
	h.RegisterHandlers(LEAVE_ROOM, h.HandleLeaveRoom)
	h.RegisterHandlers(JOIN_ROOM, h.HandleJoinRoom)

}

func (h *Hub) publishToRedis(eventType string, payload Message, roomId string) {
	redisMessage := RedisMessage{
		Event: Event{
			Type:    eventType,
			Payload: payload,
		},
		RoomId:   roomId,
		ServerId: h.serverName,
	}
	data, err := json.Marshal(redisMessage)
	if err != nil {
		logger.Errorln("Error while marshing the data", err)
		return
	}
	ch := fmt.Sprintf("chat:room:%s", roomId)
	if err := h.redisClient.Publish(h.ctx, ch, data).Err(); err != nil {
		logger.Errorln("Failed to publish To redis", err)
		return
	}
	logger.Infof("EVENT publish to redis successfully")
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
	err := h.redisClient.HSet(h.ctx, roomId)
	if err != nil {
		logger.Errorln("Errro while creating room in redis", err)
	}
	h.publishToRedis(CREATE_ROOM, event.Payload, roomId)
	logger.Infof("Room is created")

	return nil
}
func (h *Hub) HandleSendMessage(event Event, client *Client) error {
	roomId := event.Payload.RoomId
	if roomId == "" {
		return fmt.Errorf("Room ID is missing")
	}
	h.Mu.RLock()
	room, exist := h.Rooms[roomId]
	h.Mu.RUnlock()
	if !exist {
		return fmt.Errorf("Room Does not Exist with roomID %s", roomId)
	}
	message := NewMessage(client.Username, event.Payload.Content, roomId)

	h.publishToRedis(MESSAGE_RECEVIED, message, room.RoomId)
	return nil
}
func (h *Hub) HandleLeaveRoom(event Event, client *Client) error {
	roomID := event.Payload.RoomId
	if roomID == "" {
		return fmt.Errorf("missing room ID")
	}
	h.Mu.RLock()
	if _, exist := h.Rooms[roomID]; !exist {
		return fmt.Errorf("room does not exist")
	}
	h.Mu.RUnlock()
	room := h.Rooms[roomID]
	if err := room.removeClient(client); err != nil {
		logger.Errorln("Error while removing the client from Room", err)
		return err
	}
	leaveMsg := NewMessage("System", fmt.Sprintf("%s has left the Room", client.Username), roomID)
	h.publishToRedis(USER_LEFT, leaveMsg, roomID)
	logger.Infof("User %s Has Left the Room", client.Username)
	return nil
}
func (h *Hub) HandleJoinRoom(event Event, client *Client) error {
	roomId := event.Payload.RoomId
	var room *Room
	if roomId == "" {
		return fmt.Errorf("Room ID is empty")
	}
	h.Mu.RLock()
	if _, exist := h.Rooms[roomId]; !exist {
		//if not found then find it redis
		room, err := h.redisClient.HGetAll(h.ctx, fmt.Sprintf("chat:room:%s", roomId)).Result()
		logger.Infof("rooms in redis are", room)
		if err != nil {
			logger.Errorln("Error while getting the room from redis", err)
			return fmt.Errorf("room does not exist")
		}
		return fmt.Errorf("room does not exist")
	}
	h.Mu.RUnlock()
	room = h.Rooms[roomId]
	logger.Infof("room is %s", room.RoomId)
	room.addClients(client)
	joinMessage := NewMessage(client.Username, fmt.Sprintf("%s joined the room", client.Username), roomId)
	logger.Infof("total room clients %s", len(room.Clients))
	logger.Infof("NEW MESSAGE %s", joinMessage)
	h.publishToRedis(USER_JOINED, joinMessage, roomId)
	logger.Infof("User ", client.Username, " has joined the room and message is published to redis")
	return nil
}
func (h *Hub) startRedisSubscriber() {
	go func() {
		pattern := "chat:room:*"
		h.pubsub = h.redisClient.PSubscribe(h.ctx, pattern)
		defer func() {
			h.pubsub.Close()
			logger.Infof("Redis subscriber stopped")
		}()

		logger.Infof("Redis Subscriber started")
		ch := h.pubsub.Channel()
		for {
			select {
			case <-h.ctx.Done():
				logger.Infof("Context Done Called in Redis Sub")
				return
			case msg := <-ch:
				h.handleRedisMessage(msg)
			}
		}
	}()
}
func (h *Hub) handleRedisMessage(msg *redis.Message) {
	var redisMessage RedisMessage
	if err := json.Unmarshal([]byte(msg.Payload), &redisMessage); err != nil {
		logger.Errorln("Failed to UnMarshal Redis Message", err)
		return
	}

	logger.Infof("Received Redis message for room %s", redisMessage.RoomId)

	h.Mu.RLock()
	room, exist := h.Rooms[redisMessage.RoomId]
	h.Mu.RUnlock()

	if !exist {
		logger.Warn("The Room Does not Exist on Redis returning...")
		return
	}

	logger.Infof("Broadcasting Redis message to room %s", redisMessage.RoomId)
	room.Broadcast(redisMessage.Event, nil)
}
