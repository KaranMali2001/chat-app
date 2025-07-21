package main

import (
	"context"
	"os"

	"github.com/chat-app/internal/config"
	"github.com/chat-app/internal/handler"
	"github.com/chat-app/internal/hub"
	"github.com/chat-app/internal/metrics"
	"github.com/chat-app/pkg/logger"
	"github.com/chat-app/pkg/redis"
	"github.com/joho/godotenv"
	goRedis "github.com/redis/go-redis/v9"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic("ENV NOT LOADED")
	}

}
func initLogger() {
	isProd := os.Getenv("APP_ENV") == "production"
	if err := logger.Init(isProd); err != nil {
		panic("Failed To initize the Logger")
	}
	logger.Infof("Logger Initiziled successfully")
}
func initConfig() {
	config.LoadServerConfig()
	logger.Infof("Config loaded")
}

func initMetrics() {
	metrics.Init()
	logger.Infof("Metrics initialized successfully")
}
func initRedis() *goRedis.Client {
	redisConfig := config.LoadRedisConfig()
	config.LoadServerConfig()
	rds, err := redis.InitRedisCleint(redisConfig.Addr, redisConfig.Password)
	if err != nil {
		panic("Redis is not iniizlted")
	}
	if err := rds.Ping(context.Background()).Err(); err != nil {
		panic("Not able to ping Redis")
	}

	logger.Infof("Redis connection established successfully")
	return rds

}
func initHub(rds *goRedis.Client) {
	chathub = hub.NewHub(rds, config.AppConfig.Name)
	handler.SetHub(chathub)
}
