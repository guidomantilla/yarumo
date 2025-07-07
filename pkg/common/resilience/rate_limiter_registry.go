package resilience

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiterRegistry struct {
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
	limiters map[string]*rate.Limiter
}

func NewRateLimiterRegistry() *RateLimiterRegistry {
	return &RateLimiterRegistry{
		rate:     rate.Every(100 * time.Millisecond),
		burst:    5,
		limiters: make(map[string]*rate.Limiter),
	}
}

func (r *RateLimiterRegistry) Get(name string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	l, ok := r.limiters[name]
	if ok {
		return l
	}

	limiter := rate.NewLimiter(r.rate, r.burst)
	r.limiters[name] = limiter
	return limiter
}
