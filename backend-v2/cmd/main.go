package main

import (
	"context"
	"net/http"
	"os"

	"github.com/chat-app/internal/config"
	"github.com/chat-app/internal/handler"
	"github.com/chat-app/internal/metrics"
	"github.com/chat-app/pkg/logger"
	"github.com/chat-app/pkg/redis"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("ENV NOT LOADED")
	}
	isProd := os.Getenv("APP_ENV") == "production"
	if err := logger.Init(isProd); err != nil {
		panic("Failed To initize the Logger")
	}
	logger.Infof("Logger Initiziled successfully")
	redisConfig := config.LoadRedisConfig()
	rds, err := redis.InitRedisCleint(redisConfig.Addr, redisConfig.Password)
	if err != nil {
		panic("Redis is not iniizlted")
	}
	if err := rds.Ping(context.Background()).Err(); err != nil {
		panic("Not able to ping Redis")
	}
	logger.Infof("Redis connection established successfully")
	metrics.Init()
	logger.Infof("Metrics initizalied successfully")
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", metrics.InstrumentHTTP("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("SERVER IS RUNNING"))
	})))

	http.HandleFunc("/ws", handler.WebSocketUpgrader)
	config.LoadServerConfig()
	logger.Infof("Server started at PORT %s and server name is %s", config.AppConfig.Port, config.AppConfig.Name)
	err = http.ListenAndServe(config.AppConfig.Port, nil)
	if err != nil {
		panic("HTTP SERVER DID NOT START")
	}
}
