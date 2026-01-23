package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	*redis.Client
}

func NewRedis(redisURL string) (*RedisDB, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// Try simple host:port format
		addr := strings.TrimPrefix(redisURL, "redis://")
		opts = &redis.Options{
			Addr: addr,
		}
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisDB{client}, nil
}

func (r *RedisDB) Close() error {
	return r.Client.Close()
}
