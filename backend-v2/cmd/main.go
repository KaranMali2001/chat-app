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

	// Register routes on a custom mux so we can wrap with CORS
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/health", handler.HealthHandler)
	mux.HandleFunc("/api/v1/ws", handler.WebSocketUpgrader)
	mux.HandleFunc("/api/v1/create-room", handler.CreateRoom)
	mux.HandleFunc("/api/v1/room-stats", handler.GetRoomStats)

	// Apply CORS middleware to all routes
	handlerWithCORS := withCORS(mux)

	logger.Infof("Server started at PORT %s and server name is %s", config.AppConfig.Port, config.AppConfig.Name)
	if err := http.ListenAndServe(config.AppConfig.Port, handlerWithCORS); err != nil {
		panic("HTTP SERVER DID NOT START")
	}
}

// withCORS wraps an HTTP handler to allow requests from localhost:5173
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow cookies/auth headers

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
