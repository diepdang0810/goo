package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go1/internal/modules/user/domain"
	"go1/internal/modules/user/infrastructure/caching/mapper"
	"go1/internal/modules/user/infrastructure/caching/model"
	"go1/pkg/redis"
)

type redisUserCache struct {
	client *redis.RedisClient
}

func NewRedisUserCache(client *redis.RedisClient) domain.UserCache {
	return &redisUserCache{client: client}
}

func (c *redisUserCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	key := fmt.Sprintf("user:%d", id)
	val, err := c.client.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redisUserCache.Get redis: %w", err)
	}

	var cachingModel model.UserCachingModel
	if err := json.Unmarshal([]byte(val), &cachingModel); err != nil {
		return nil, fmt.Errorf("redisUserCache.Get unmarshal: %w", err)
	}

	return mapper.ToDomain(&cachingModel), nil
}

func (c *redisUserCache) Set(ctx context.Context, user *domain.User) error {
	key := fmt.Sprintf("user:%d", user.ID)
	
	cachingModel := mapper.ToCachingModel(user)
	data, err := json.Marshal(cachingModel)
	if err != nil {
		return fmt.Errorf("redisUserCache.Set marshal: %w", err)
	}

	if err := c.client.Client.Set(ctx, key, data, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("redisUserCache.Set redis: %w", err)
	}
	return nil
}

func (c *redisUserCache) Delete(ctx context.Context, id int64) error {
	key := fmt.Sprintf("user:%d", id)
	if err := c.client.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redisUserCache.Delete: %w", err)
	}
	return nil
}
