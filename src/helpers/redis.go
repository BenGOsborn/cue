package helpers

import "github.com/redis/go-redis/v9"

func NewRedis(redisUrl string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	return client, nil
}
