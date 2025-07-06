package handler

import (
	"net/http"
	"slices"

	"github.com/chat-app/internal/config"
	"github.com/chat-app/pkg/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: checkOrigin,
}

// var AllowedOrigins = config.LoadServerConfig().AllowedOrigins //this wont work because this is package level vairable and it gets initizaled before main hence godotenv() func didnt get call and hence env is not loaded

func WebSocketUpgrader(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	logger.Infof("Username is %s", username)

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		logger.Errorln("Error while upgrading the websocket conn", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cant upgrade websocket connection"))
		return
	}
	defer conn.Close()
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Errorln("error while reading the message ", err)

			break

		}
		logger.Infof("Message recevived %s", msg)
		if err := conn.WriteMessage(msgType, msg); err != nil {
			logger.Errorln("error while Writing back to client", err)
		}
	}
}
func checkOrigin(r *http.Request) bool {

	origin := r.Header.Get("Origin")
	logger.Infof("Origin %s", origin)

	return origin != "" && slices.Contains(config.AppConfig.AllowedOrigins, (origin))
}
