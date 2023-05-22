// Package pkg ...
package pkg

import (
	goredis "github.com/redis/go-redis/v9"
)

// RedisClient is a wrapper of redis client
type RedisClient struct {
	goredis.UniversalClient
}

// RedisClientConfig is the config of redis client
type RedisClientConfig struct {
	Addr string
	DB   int
}

// NewRedisClient returns a new redis client
func NewRedisClient(cfg RedisClientConfig) *RedisClient {
	client := goredis.NewClient(&goredis.Options{
		Addr: cfg.Addr,
		DB:   cfg.DB,
	})
	return &RedisClient{client}
}
