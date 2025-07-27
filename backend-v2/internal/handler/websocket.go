package handler

import (
	"net/http"
	"slices"
	"time"

	"github.com/chat-app/internal/config"
	"github.com/chat-app/internal/hub"
	"github.com/chat-app/internal/metrics"
	"github.com/chat-app/pkg/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: checkOrigin,
}
var chathub *hub.Hub

func SetHub(h *hub.Hub) {
	chathub = h
}

// var AllowedOrigins = config.LoadServerConfig().AllowedOrigins //this wont work because this is package level vairable and it gets initizaled before main hence godotenv() func didnt get call and hence env is not loaded

func WebSocketUpgrader(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	roomId := r.URL.Query().Get("roomid")
	if username == "" || roomId == "" {
		http.Error(w, "Missing username or roomId", http.StatusBadRequest)
		return
	}
	logger.Infof("Username is %s", username)
	logger.Infof("Roomid is %s", roomId)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		logger.Errorln("Error while upgrading the websocket conn", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cant upgrade websocket connection"))
		return
	}
	client := hub.NewClient(username, conn, chathub)
	event := hub.Event{
		Type: hub.JOIN_ROOM,
		Payload: hub.Message{
			RoomId: roomId,
			Sender: username,
			Time:   time.Now().String(),
		},
	}
	if err := chathub.ProcessEvent(event, client); err != nil {
		logger.Errorln("Error while joining room:", err)
		return
	}
	go client.ReadMessage()
	go client.WriteMessage()
	metrics.IncrementActiveConnections()
	<-client.Ctx.Done()
	defer metrics.DecreamentActiveConnections()
	defer conn.Close()
	// Handle leave room event when client disconnects
	leaveEvent := hub.Event{
		Type: hub.LEAVE_ROOM,
		Payload: hub.Message{
			RoomId: roomId,
			Sender: username,
			Time:   time.Now().String(),
		},
	}

	if err := chathub.ProcessEvent(leaveEvent, client); err != nil {
		logger.Errorln("Error while leaving room:", err)
	}

	logger.Infof("Client %s disconnected from room %s", username, roomId)
}
func checkOrigin(r *http.Request) bool {

	origin := r.Header.Get("Origin")

	return origin != "" && slices.Contains(config.AppConfig.AllowedOrigins, (origin))
}
