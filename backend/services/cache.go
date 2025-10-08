package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheService struct {
	redis *redis.Client
}

func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{redis: redis}
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, jsonData, expiration).Err()
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}

func (c *CacheService) GetOrSet(ctx context.Context, key string, dest interface{},
	fetchFunc func() (interface{}, error), expiration time.Duration) error {

	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err != redis.Nil {
		return err
	}

	data, err := fetchFunc()
	if err != nil {
		return err
	}

	err = c.Set(ctx, key, data, expiration)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data.(string)), dest)
}

func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.redis.Exists(ctx, key).Result()
	return result > 0, err
}

func (c *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	return c.redis.Incr(ctx, key).Result()
}

func (c *CacheService) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	return c.redis.Expire(ctx, key, expiration).Err()
}
