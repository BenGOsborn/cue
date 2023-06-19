package utils

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

// Format a redis key
func FormatKey(parts ...string) string {
	return strings.Join(parts, ":")
}

// Create a new redis instance
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
func (r *Redis) Set(key string, value string, expiry time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiry).Err()
}

// Retrieve a key
func (r *Redis) Get(key string) string {
	return r.client.Get(r.ctx, key).Val()
}

// Remove a key
func (r *Redis) Remove(key string) {
	r.client.Del(r.ctx, key)
}
