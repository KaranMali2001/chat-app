package redis

import (
	"github.com/redis/go-redis/v9"
)

func InitRedisCleint(addr string, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	return client, nil
}
