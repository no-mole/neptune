package cache

import (
	"context"
	"errors"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type MemLruCache struct {
	instance *lru.Cache
}

func NewMemLruCache(_ context.Context, size int) (Cache, error) {
	if size <= 0 || size > 2048 {
		return nil, errors.New("unsupported size")
	}
	instance, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &MemLruCache{
		instance: instance,
	}, nil
}

type expireValue struct {
	ctx          context.Context
	enableExpire bool
	expiredTime  time.Time
	value        interface{}
}

var ErrorValueExpired = errors.New("cached value expired")
var ErrorValueExpiredOrNotExist = errors.New("cached value expired or not exist")

func (e *expireValue) Value() (interface{}, error) {
	if !e.enableExpire || time.Now().Before(e.expiredTime) {
		return e.value, nil
	}
	return nil, ErrorValueExpired
}

func (m *MemLruCache) Get(_ context.Context, key string) (interface{}, error) {
	value, ok := m.instance.Get(key)
	if !ok {
		return nil, ErrorValueExpiredOrNotExist
	}
	return value.(*expireValue).Value()
}

func (m *MemLruCache) Set(ctx context.Context, key string, value interface{}) error {
	m.instance.Add(key, &expireValue{
		ctx:          ctx,
		enableExpire: false,
		value:        value,
	})
	return nil
}

func (m *MemLruCache) SetEx(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	m.instance.Add(key, &expireValue{
		ctx:          ctx,
		enableExpire: true,
		expiredTime:  time.Now().Add(expire),
		value:        value,
	})
	return nil
}

func (m *MemLruCache) Delete(_ context.Context, key string) (bool, error) {
	m.instance.Remove(key)
	return true, nil
}

func (m *MemLruCache) Exist(_ context.Context, key string) (bool, error) {
	_, ok := m.instance.Get(key)
	return ok, nil
}

var _ Cache = &MemLruCache{}
