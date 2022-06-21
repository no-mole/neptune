package ratelimit

import (
	"sync"
	"time"
)

type Bucket struct {
	clock           Clock
	startTime       time.Time
	cap             int64
	quantum         int64
	interval        time.Duration
	mu              sync.Mutex
	availableTokens int64
	latestTick      int64
}

func NewBucket(interval time.Duration, cap int64) *Bucket {
	if interval <= 0 {
		panic("token Bucket fill interval <= 0")
	}
	if cap <= 0 {
		panic("token Bucket cap <= 0")
	}

	return &Bucket{
		clock:           &defaultClock{},
		startTime:       time.Now(),
		latestTick:      0,
		interval:        interval,
		cap:             cap,
		quantum:         1,
		availableTokens: cap,
	}
}

func (b *Bucket) TakeAvailable(count int64) int64 {
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

func (b *Bucket) adjustAvailableTokens(now time.Time) {
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

type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

type defaultClock struct {
}

func (dc *defaultClock) Now() time.Time {
	return time.Now()
}

func (dc *defaultClock) Sleep(d time.Duration) {
	time.Sleep(d)
}
