package redis

import (
	"context"
	"time"

	"go1/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func New(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger.Log.Info("Connected to Redis", logger.Field{Key: "addr", Value: addr})
	return &RedisClient{Client: client}, nil
}

func (r *RedisClient) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}
