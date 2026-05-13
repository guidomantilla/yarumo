package resilience

import (
	"context"
	"sync"

	"golang.org/x/time/rate"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// rateLimiter wraps a *rate.Limiter behind the RateLimiter interface so
// callers see only this package's error type.
type rateLimiter struct {
	limiter *rate.Limiter
}

// newRateLimiter builds a rateLimiter from the configured options.
func newRateLimiter(opts *Options) *rateLimiter {
	cassert.NotNil(opts, "options is nil")

	return &rateLimiter{limiter: rate.NewLimiter(opts.rateLimit(), opts.rateBurst)}
}

// Allow reports whether a token is available right now without blocking.
func (l *rateLimiter) Allow() bool {
	cassert.NotNil(l, "rate limiter is nil")
	cassert.NotNil(l.limiter, "underlying limiter is nil")

	return l.limiter.Allow()
}

// Wait blocks until a token is available or ctx is canceled.
func (l *rateLimiter) Wait(ctx context.Context) error {
	cassert.NotNil(l, "rate limiter is nil")
	cassert.NotNil(l.limiter, "underlying limiter is nil")

	err := validateWait(ctx)
	if err != nil {
		return ErrRateLimiterWait(err)
	}

	err = l.limiter.Wait(ctx)
	if err != nil {
		return ErrRateLimiterWait(cerrs.Wrap(err))
	}

	return nil
}

// rateLimiterRegistry is the lazy, goroutine-free registry implementation.
type rateLimiterRegistry struct {
	lock     sync.RWMutex
	limiters map[string]*rateLimiter
}

// NewRateLimiterRegistry creates a new lazy RateLimiterRegistry.
func NewRateLimiterRegistry() RateLimiterRegistry {
	return &rateLimiterRegistry{
		limiters: make(map[string]*rateLimiter),
	}
}

// Get returns the limiter registered under name, lazy-creating it with
// default options if none exists.
//
// An empty name is accepted by Get (it maps to a single shared default
// limiter). Use Use(name) to validate name explicitly.
func (r *rateLimiterRegistry) Get(name string) RateLimiter {
	cassert.NotNil(r, "registry is nil")

	r.lock.RLock()
	existing, ok := r.limiters[name]
	r.lock.RUnlock()

	if ok {
		return existing
	}

	return r.loadOrCreate(name)
}

// Use replaces (or creates) the named entry with a limiter built from the
// supplied options. It returns a domain error when name is empty.
func (r *rateLimiterRegistry) Use(name string, opts ...Option) error {
	cassert.NotNil(r, "registry is nil")

	err := validateName(name)
	if err != nil {
		return ErrRegistryUse(err)
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.limiters[name] = newRateLimiter(NewOptions(opts...))

	return nil
}

// loadOrCreate finalizes the lazy-create path under the write lock. The
// recheck protects against a concurrent writer that inserted between the
// RUnlock and Lock calls in Get.
func (r *rateLimiterRegistry) loadOrCreate(name string) RateLimiter {
	r.lock.Lock()
	defer r.lock.Unlock()

	existing, ok := r.limiters[name]
	if ok {
		return existing
	}

	created := newRateLimiter(NewOptions())
	r.limiters[name] = created

	return created
}
