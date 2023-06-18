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
func (r *Redis) Set(prefix string, key string, value string, expiry time.Duration) error {
	return r.client.Set(r.ctx, formatKey(prefix, key), value, expiry).Err()
}

// Retrieve a key
func (r *Redis) Get(prefix string, key string) string {
	return r.client.Get(r.ctx, formatKey(prefix, key)).Val()
}

// Remove a key
func (r *Redis) Remove(prefix string, key string) {
	r.client.Del(r.ctx, formatKey(prefix, key))
}
