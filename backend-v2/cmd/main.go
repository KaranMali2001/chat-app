package main

import (
	"context"
	"os"

	"github.com/chat-app/internal/config"
	"github.com/chat-app/pkg/logger"
	"github.com/chat-app/pkg/redis"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
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
	zap.L().Info("Logger Initiziled successfully")
	redisConfig := config.LoadRedisConfig()
	rds, err := redis.InitRedisCleint(redisConfig.Addr, redisConfig.Password)
	if err != nil {
		panic("Redis is not iniizlted")
	}
	if err := rds.Ping(context.Background()).Err(); err != nil {
		panic("Not able to ping Redis")
	}
	zap.L().Info("Redis connection established successfully")

}
