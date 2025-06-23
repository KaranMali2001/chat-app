package routes

import (
	"net/http"

	"github.com/chat-app/websocket"
	"go.uber.org/zap"
)

func RegisterRoutes(logger *zap.SugaredLogger) {
	manager := websocket.NewManager(logger)
	http.HandleFunc("/ws", manager.ServeWs)
}
