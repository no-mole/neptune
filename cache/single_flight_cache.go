package cache

import (
	"context"
	"sync"
	"time"
)

func NewSingleFlightCache(ctx context.Context, cacheTime time.Duration, fn SingleFlightFunc) (Cache, error) {
	return &SingleFlightCache{
		ctx:        ctx,
		cacheTime:  cacheTime,
		cachedTime: time.Now().Add(-cacheTime),
		fn:         fn,
		mux:        sync.Mutex{},
	}, nil
}

type SingleFlightFunc func(ctx context.Context) (data interface{}, err error)

type SingleFlightCache struct {
	ctx         context.Context
	data        interface{}
	err         error
	cacheTime   time.Duration
	cachedTime  time.Time
	fn          SingleFlightFunc
	mux         sync.Mutex
	lastGetTime time.Time
}

func (s *SingleFlightCache) checker() {
	go func() {
		ticker := time.NewTicker(s.cacheTime)
		for {
			select {
			case <-ticker.C:
				if time.Since(s.lastGetTime) > s.cacheTime*2 {
					//超过两倍cacheTime没有访问,释放data以回收内存
					s.mux.Lock()
					s.data = nil
					s.err = nil
					s.mux.Unlock()
				}
			case <-s.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *SingleFlightCache) Get(ctx context.Context, _ string) (interface{}, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if time.Since(s.cachedTime) > s.cacheTime {
		s.data, s.err = s.fn(ctx)
		s.cachedTime = time.Now()
	}
	s.lastGetTime = time.Now()
	return s.data, s.err
}

func (s *SingleFlightCache) Set(_ context.Context, _ string, value interface{}) error {
	s.mux.Lock()
	s.cachedTime = time.Now()
	s.data = value
	s.mux.Unlock()
	return nil
}

func (s *SingleFlightCache) SetEx(_ context.Context, _ string, value interface{}, expire time.Duration) error {
	s.mux.Lock()
	//重置缓存时间
	s.cacheTime = expire
	s.cachedTime = time.Now()
	s.data = value
	s.err = nil
	s.mux.Unlock()
	return nil
}

func (s *SingleFlightCache) Delete(_ context.Context, _ string) (bool, error) {
	s.mux.Lock()
	//设置缓存时间，让缓存过期
	s.cachedTime = time.Now().Add(s.cacheTime)
	s.data = nil
	s.err = nil
	s.mux.Unlock()
	return true, nil
}

func (s *SingleFlightCache) Exist(_ context.Context, _ string) (bool, error) {
	//缓存没过期
	s.mux.Lock()
	defer s.mux.Unlock()
	return time.Since(s.cachedTime) < s.cacheTime, nil
}

var _ Cache = &SingleFlightCache{}
