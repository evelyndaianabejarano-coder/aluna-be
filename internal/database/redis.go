package database

import (
	"context"

	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
