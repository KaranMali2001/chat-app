package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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

	// Check if room exists in Redis first (for distributed environment)
	roomKey := fmt.Sprintf("chat:room:%s", roomId)
	exists, err := h.redisClient.Exists(h.ctx, roomKey).Result()
	if err != nil {
		logger.Errorln("Error checking room existence in Redis", err)
		return fmt.Errorf("failed to check room existence")
	}

	if exists > 0 {
		return fmt.Errorf("room already exists, try to join the room")
	}

	h.Mu.Lock()
	defer h.Mu.Unlock()

	// Create room locally
	newRoom := createRoom(roomId)
	h.Rooms[roomId] = newRoom

	// Store room in Redis for distributed access
	roomData := map[string]interface{}{
		"room_id":    roomId,
		"created_by": client.Username,
		"server_id":  h.serverName,
		"created_at": fmt.Sprintf("%d", time.Now().Unix()),
	}

	if err := h.redisClient.HSet(h.ctx, roomKey, roomData).Err(); err != nil {
		logger.Errorln("Error while creating room in Redis", err)
		// Remove from local storage if Redis fails
		delete(h.Rooms, roomId)
		return fmt.Errorf("failed to create room in Redis")
	}

	// Set room expiration (optional: 24 hours)
	h.redisClient.Expire(h.ctx, roomKey, 24*time.Hour)

	h.publishToRedis(CREATE_ROOM, event.Payload, roomId)
	logger.Infof("Room %s created successfully", roomId)

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

	// Broadcast locally first
	broadcastEvent := Event{
		Type:    MESSAGE_RECEVIED,
		Payload: message,
	}
	room.Broadcast(broadcastEvent, nil)

	// Then publish to Redis for other servers
	h.publishToRedis(MESSAGE_RECEVIED, message, room.RoomId)
	return nil
}
func (h *Hub) HandleLeaveRoom(event Event, client *Client) error {
	roomID := event.Payload.RoomId
	if roomID == "" {
		return fmt.Errorf("missing room ID")
	}

	h.Mu.Lock()
	defer h.Mu.Unlock()

	room, exists := h.Rooms[roomID]
	if !exists {
		return fmt.Errorf("room does not exist")
	}

	if err := room.removeClient(client); err != nil {
		logger.Errorln("Error while removing client from room", err)
		return err
	}

	// Update Redis with current client count
	clientCountKey := fmt.Sprintf("chat:room:%s:clients:%s", roomID, h.serverName)
	h.redisClient.Set(h.ctx, clientCountKey, len(room.Clients), time.Hour)

	// If room is empty, optionally clean it up
	if len(room.Clients) == 0 {
		logger.Infof("Room %s is empty, cleaning up locally", roomID)
		delete(h.Rooms, roomID)
		// Remove client count from Redis
		h.redisClient.Del(h.ctx, clientCountKey)
	}

	leaveMsg := NewMessage("System", fmt.Sprintf("%s has left the room", client.Username), roomID)
	h.publishToRedis(USER_LEFT, leaveMsg, roomID)
	logger.Infof("User %s has left room %s", client.Username, roomID)
	return nil
}
func (h *Hub) HandleJoinRoom(event Event, client *Client) error {
	roomId := event.Payload.RoomId
	if roomId == "" {
		return fmt.Errorf("Room ID is empty")
	}

	// Store room ID in client for reconnection
	client.mu.Lock()
	client.roomID = roomId
	client.mu.Unlock()

	h.Mu.Lock()
	defer h.Mu.Unlock()

	// Check if room exists locally
	room, exists := h.Rooms[roomId]
	if !exists {
		// Try to load room from Redis
		roomKey := fmt.Sprintf("chat:room:%s", roomId)
		roomData, err := h.redisClient.HGetAll(h.ctx, roomKey).Result()
		if err != nil && err != redis.Nil {
			logger.Errorln("Error getting room from Redis: %v", err)
			return fmt.Errorf("failed to check room existence")
		}

		if len(roomData) == 0 {
			// Room doesn't exist, create it
			room = &Room{
				RoomId:  roomId,
				Clients: make(map[string]*Client),
			}
			h.Rooms[roomId] = room
			
			// Save room to Redis
			roomJSON, err := json.Marshal(map[string]interface{}{
				"id":      roomId,
				"created": time.Now().Unix(),
			})
			if err != nil {
				logger.Errorln("Error marshaling room data: %v", err)
				return fmt.Errorf("failed to create room")
			}
			
			if err := h.redisClient.HSet(h.ctx, roomKey, roomJSON).Err(); err != nil {
				logger.Errorln("Error saving room to Redis: %v", err)
				// Continue even if Redis save fails, as we have it in memory
			}
			
			logger.Infof("Created new room: %s", roomId)
		} else {
			// Room exists in Redis but not locally, create it
			room = &Room{
				RoomId:  roomId,
				Clients: make(map[string]*Client),
			}
			h.Rooms[roomId] = room
			logger.Infof("Loaded room from Redis: %s", roomId)
		}
	}

	// Add client to room
	if err := room.addClients(client); err != nil {
		logger.	Errorln("Error adding client to room: %v", err)
		return fmt.Errorf("failed to join room")
	}

	joinMessage := NewMessage(client.Username, fmt.Sprintf("%s joined the room", client.Username), roomId)
	logger.Infof("Total room clients: %d", len(room.Clients))

	// Update Redis with current server's client count
	clientCountKey := fmt.Sprintf("chat:room:%s:clients:%s", roomId, h.serverName)
	if err := h.redisClient.Set(h.ctx, clientCountKey, len(room.Clients), time.Hour).Err(); err != nil {
		logger.Errorln("Error updating client count in Redis: %v", err)
	}

	h.publishToRedis(USER_JOINED, joinMessage, roomId)
	logger.Infof("User %s has joined room %s", client.Username, roomId)
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

	logger.Infof("Received Redis message for room %s from server %s", redisMessage.RoomId, redisMessage.ServerId)

	h.Mu.RLock()
	room, exist := h.Rooms[redisMessage.RoomId]
	h.Mu.RUnlock()

	if !exist {
		logger.Infof("Room %s does not exist locally, skipping broadcast", redisMessage.RoomId)
		return
	}

	logger.Infof("Broadcasting Redis message to room %s with %d clients", redisMessage.RoomId, len(room.Clients))
	room.Broadcast(redisMessage.Event, nil)
}

// Cleanup method to be called on server shutdown
func (h *Hub) Cleanup() {
	logger.Infof("Starting hub cleanup...")

	h.cancel() // Cancel context to stop Redis subscriber

	h.Mu.Lock()
	defer h.Mu.Unlock()

	// Clean up client counts for all rooms on this server
	for roomId := range h.Rooms {
		clientCountKey := fmt.Sprintf("chat:room:%s:clients:%s", roomId, h.serverName)
		h.redisClient.Del(h.ctx, clientCountKey)
	}

	// Close pubsub connection
	if h.pubsub != nil {
		h.pubsub.Close()
	}

	logger.Infof("Hub cleanup completed")
}
