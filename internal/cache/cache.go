package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) Cache {
	return &RedisCache{
		client: client,
	}
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil // Key không tồn tại
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		err := c.client.Del(ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	return iter.Err()
}
