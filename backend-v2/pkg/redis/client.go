package redis

import (
	"crypto/tls"

	"github.com/redis/go-redis/v9"
)

func InitRedisCleint(redisHost, username, password string, ssl bool) (*redis.Client, error) {
	options := &redis.Options{
		Addr:     redisHost,
		Username: username,
		Password: password,
		DB:       0,
	}
	if ssl {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	client := redis.NewClient(options)
	return client, nil
}
