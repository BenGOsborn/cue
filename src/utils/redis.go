package utils

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

func formatKey(prefix string, key string) string {
	return prefix + ":" + key
}

func NewRedis(ctx context.Context, redisUrl string) (*Redis, error) {
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	return &Redis{client: client, ctx: ctx}, nil
}

// Close the connection
func (redis *Redis) Close() {
	redis.client.Close()
}

// Add a new key
func (redis *Redis) Set(prefix string, key string, value string, expiry time.Duration) {
	redis.client.Set(redis.ctx, formatKey(prefix, key), value, expiry)
}

// Retrieve a key
func (redis *Redis) Get(prefix string, key string) string {
	return redis.client.Get(redis.ctx, formatKey(prefix, key)).Val()
}

// Remove a key
func (redis *Redis) Remove(prefix string, key string) {
	redis.client.Del(redis.ctx, formatKey(prefix, key))
}
