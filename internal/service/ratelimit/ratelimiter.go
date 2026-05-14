package ratelimit

import (
	"context"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type memoryRateLimiter struct {
	mu      sync.Mutex
	limits  map[string]*bucket
	max     int
	window  time.Duration
}

type bucket struct {
	count     int
	resetTime time.Time
}

func NewMemoryRateLimiter(max int, window time.Duration) RateLimiter {
	return &memoryRateLimiter{
		limits: make(map[string]*bucket),
		max:    max,
		window: window,
	}
}

func (r *memoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, exists := r.limits[key]

	if !exists || now.After(b.resetTime) {
		r.limits[key] = &bucket{
			count:     1,
			resetTime: now.Add(r.window),
		}
		return true, nil
	}

	if b.count >= r.max {
		return false, nil
	}

	b.count++
	return true, nil
}
