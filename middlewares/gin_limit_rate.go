package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

func LimitRateWithBucket(interval time.Duration, cap int64) gin.HandlerFunc {
	bucket := newBucket(interval, cap)
	return func(ctx *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			ctx.String(http.StatusForbidden, "rate limit")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

type bucket struct {
	clock           clock
	startTime       time.Time
	cap             int64
	quantum         int64
	interval        time.Duration
	mu              sync.Mutex
	availableTokens int64
	latestTick      int64
}

func newBucket(interval time.Duration, cap int64) *bucket {
	if interval <= 0 {
		panic("token bucket fill interval <= 0")
	}
	if cap <= 0 {
		panic("token bucket cap <= 0")
	}
	return &bucket{
		clock: clock{
			Now: func() time.Time {
				return time.Now()
			},
			Sleep: func(d time.Duration) {
				time.Sleep(d)
			},
		},
		startTime:       time.Now(),
		latestTick:      0,
		interval:        interval,
		cap:             cap,
		quantum:         1,
		availableTokens: cap,
	}
}

func (b *bucket) TakeAvailable(count int64) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	if count <= 0 {
		return 0
	}

	b.adjustAvailableTokens(b.clock.Now())

	if b.availableTokens <= 0 {
		return 0
	}

	if count > b.availableTokens {
		count = b.availableTokens
	}
	b.availableTokens -= count
	return count
}

func (b *bucket) adjustAvailableTokens(now time.Time) {
	tick := int64(now.Sub(b.startTime) / b.interval)
	lastTick := b.latestTick
	b.latestTick = tick

	if b.availableTokens >= b.cap {
		return
	}
	b.availableTokens += (tick - lastTick) * b.quantum

	if b.availableTokens > b.cap {
		b.availableTokens = b.cap
	}
	return
}

type clock struct {
	Now   func() time.Time
	Sleep func(d time.Duration)
}
