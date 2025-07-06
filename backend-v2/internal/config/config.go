package config

import (
	"os"
	"strings"
)

type ServerConfig struct {
	Name           string
	Port           string
	AllowedOrigins []string
}

var AppConfig ServerConfig

func LoadServerConfig() {
	AppConfig = ServerConfig{
		Name:           os.Getenv("SERVER_NAME"),
		Port:           os.Getenv("SERVER_PORT"),
		AllowedOrigins: strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
	}
}
