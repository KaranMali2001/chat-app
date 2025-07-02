package config

import "os"

type RedisConfig struct {
	Addr     string
	Password string
}

// LoadRedisConfig returns Redis config for local/dev or prod environment
func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
}
