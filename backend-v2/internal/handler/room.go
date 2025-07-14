package handler

import (
	"math/rand"
	"net/http"

	"github.com/chat-app/internal"
	"github.com/chat-app/internal/hub"

	"github.com/chat-app/pkg/logger"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func CreateRoom(w http.ResponseWriter, r *http.Request) {

	roomId := make([]byte, 5)
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is Required", http.StatusBadRequest)
		return
	}
	for i := range roomId {

		roomId[i] = charset[rand.Intn(len(charset))]
	}
	logger.Infof("Room ID is ", string(roomId))
	event := hub.Event{
		Type: hub.CREATE_ROOM,
		Payload: hub.Message{
			RoomId: string(roomId),
		},
	}
	if err := chathub.ProcessEvent(event, &hub.Client{}); err != nil {

		logger.Errorln("Error while creating Room", err)
		http.Error(w, "failed to create the room", http.StatusInternalServerError)
		return
	}

	internal.SendJson(true, map[string]interface{}{
		"message": "Room created successfuly",
		"data":    string(roomId),
	}, nil, w)

}
