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

	mux := http.NewServeMux()

	// Register routes
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/health", handler.HealthHandler)
	mux.HandleFunc("/api/v1/ws", handler.WebSocketUpgrader)
	mux.HandleFunc("/api/v1/create-room", handler.CreateRoom)
	mux.HandleFunc("/api/v1/room-stats", handler.GetRoomStats)

	// Wrap the mux with CORS middleware
	logger.Infof("Server started at PORT %s and server name is %s", config.AppConfig.Port, config.AppConfig.Name)
	err := http.ListenAndServe(config.AppConfig.Port, corsMiddleware(mux))
	if err != nil {
		panic("HTTP SERVER DID NOT START")
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
