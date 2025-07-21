package main

import (
	"net/http"

	"github.com/chat-app/internal/config"
	"github.com/chat-app/internal/handler"
	"github.com/chat-app/internal/hub"

	"github.com/chat-app/pkg/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var chathub *hub.Hub

func main() {
	loadEnv()
	initLogger()

	initConfig()
	initMetrics()
	rds := initRedis()
	initHub(rds)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", handler.HealthHandler)
	http.HandleFunc("/ws", handler.WebSocketUpgrader)
	http.HandleFunc("/api/v1/create-room", handler.CreateRoom)
	http.HandleFunc("/api/v1/room-stats", handler.GetRoomStats)

	logger.Infof("Server started at PORT %s and server name is %s", config.AppConfig.Port, config.AppConfig.Name)
	err := http.ListenAndServe(config.AppConfig.Port, nil)
	if err != nil {
		panic("HTTP SERVER DID NOT START")
	}
}
