package main

import (
	"net/http"

	"github.com/chat-app/logger"
	"github.com/chat-app/routes"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {

	logger := setUpLogger()
	err := godotenv.Load()
	if err != nil {
		logger.Errorw("Error loading .env file", err)
	}
	setUpServer(logger)

}
func setUpLogger() *zap.SugaredLogger {
	return logger.GetLogger()
}
func setUpServer(logger *zap.SugaredLogger) {

	routes.RegisterRoutes(logger)
	logger.Infoln("server running at 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Errorw("error while starting server", err)
	}

}
