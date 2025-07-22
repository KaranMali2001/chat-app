package config

import "os"

type RedisConfig struct {
	Host     string
	User     string
	Password string
}

// LoadRedisConfig returns Redis config for local/dev or prod environment
func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		User:     os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
}
