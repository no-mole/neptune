package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	SetEx(ctx context.Context, key string, value interface{}, expire time.Duration) error
	Delete(ctx context.Context, key string) (bool, error)
	Exist(ctx context.Context, key string) (bool, error)
}
