package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (s RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	return s.client.Get(ctx, key).Bytes()
}
func (s RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return s.client.Set(ctx, key, value, 0).Err()
}
func (s RedisCache) SetEx(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return s.client.Set(ctx, key, value, expire).Err()
}
func (s RedisCache) Delete(ctx context.Context, key string) (bool, error) {
	err := s.client.Del(ctx, key).Err()
	return err == nil, err
}
func (s RedisCache) Exist(ctx context.Context, key string) (bool, error) {
	_, err := s.client.Get(ctx, key).Bytes()
	return err == nil, err
}

var _ Cache = &RedisCache{}
