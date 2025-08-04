package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chat-app/internal/config"
	"github.com/chat-app/internal/handler"
	"github.com/chat-app/internal/hub"
	"github.com/chat-app/internal/metrics"
	"github.com/chat-app/pkg/logger"
	"github.com/chat-app/pkg/redis"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"

	goRedis "github.com/redis/go-redis/v9"
)

var isProd bool

func loadEnv() {
	appEnv := os.Getenv("APP_ENV")

	var filename string
	if appEnv == "production" {
		filename = ".env.prod"
	} else {
		filename = ".env.dev"
	}
	fmt.Println("xyz", filename)
	err := godotenv.Load(filename)
	if err != nil {
		fmt.Println("ERROr", err)
		panic("ENV NOT LOADED")
	}

}
func initLogger() {
	isProd = os.Getenv("APP_ENV") == "production"
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
	reg := prometheus.WrapRegistererWith(
		prometheus.Labels{"server_name": config.AppConfig.Name},
		prometheus.DefaultRegisterer,
	)
	metrics.Init(reg)
	logger.Infof("Metrics initialized successfully")
}
func initRedis() *goRedis.Client {
	redisConfig := config.LoadRedisConfig()
	config.LoadServerConfig()
	rds, err := redis.InitRedisCleint(redisConfig.Host, redisConfig.User, redisConfig.Password, isProd)
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
