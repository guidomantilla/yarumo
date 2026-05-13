package resilience

import (
	"context"
	"sync"

	"github.com/sony/gobreaker"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// circuitBreaker wraps a *gobreaker.CircuitBreaker behind the CircuitBreaker
// interface so callers see only this package's State / error types.
type circuitBreaker struct {
	breaker *gobreaker.CircuitBreaker
}

// newCircuitBreaker builds a circuitBreaker from the configured options.
func newCircuitBreaker(name string, opts *Options) *circuitBreaker {
	cassert.NotNil(opts, "options is nil")

	return &circuitBreaker{breaker: gobreaker.NewCircuitBreaker(settingsFor(name, opts))}
}

// Execute invokes fn through the breaker. If the breaker is open the call is
// rejected with ErrCircuitBreakerOpen. If it is half-open and the probe budget
// is exhausted the call is rejected with ErrCircuitBreakerTooManyRequests.
// Failures returned by fn are recorded and may trip the breaker.
func (c *circuitBreaker) Execute(ctx context.Context, fn func() (any, error)) (any, error) {
	cassert.NotNil(c, "circuit breaker is nil")
	cassert.NotNil(c.breaker, "underlying breaker is nil")

	err := validateExecute(ctx, fn)
	if err != nil {
		return nil, ErrCircuitBreakerExecute(err)
	}

	wrapped := func() (any, error) {
		// Surface ctx cancellation as the failure before invoking fn so the
		// breaker records the cancellation as a failure.
		err := ctx.Err()
		if err != nil {
			return nil, cerrs.Wrap(err)
		}

		return fn()
	}

	result, err := c.breaker.Execute(wrapped)
	if err != nil {
		return nil, ErrCircuitBreakerExecute(translateBreakerError(err))
	}

	return result, nil
}

// State returns the current operating state of the breaker.
func (c *circuitBreaker) State() State {
	cassert.NotNil(c, "circuit breaker is nil")
	cassert.NotNil(c.breaker, "underlying breaker is nil")

	return fromGobreakerState(c.breaker.State())
}

// circuitBreakerRegistry is the lazy, goroutine-free registry implementation.
type circuitBreakerRegistry struct {
	lock     sync.RWMutex
	breakers map[string]*circuitBreaker
}

// NewCircuitBreakerRegistry creates a new lazy CircuitBreakerRegistry.
func NewCircuitBreakerRegistry() CircuitBreakerRegistry {
	return &circuitBreakerRegistry{
		breakers: make(map[string]*circuitBreaker),
	}
}

// Get returns the breaker registered under name, lazy-creating it with
// default options if none exists.
//
// An empty name is accepted by Get (it maps to a single shared default
// breaker). Use Use(name) to validate name explicitly.
func (r *circuitBreakerRegistry) Get(name string) CircuitBreaker {
	cassert.NotNil(r, "registry is nil")

	r.lock.RLock()
	existing, ok := r.breakers[name]
	r.lock.RUnlock()

	if ok {
		return existing
	}

	return r.loadOrCreate(name)
}

// Use replaces (or creates) the named entry with a breaker built from the
// supplied options. It returns a domain error when name is empty.
func (r *circuitBreakerRegistry) Use(name string, opts ...Option) error {
	cassert.NotNil(r, "registry is nil")

	err := validateName(name)
	if err != nil {
		return ErrRegistryUse(err)
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.breakers[name] = newCircuitBreaker(name, NewOptions(opts...))

	return nil
}

// loadOrCreate finalizes the lazy-create path under the write lock. The
// recheck protects against a concurrent writer that inserted between the
// RUnlock and Lock calls in Get.
func (r *circuitBreakerRegistry) loadOrCreate(name string) CircuitBreaker {
	r.lock.Lock()
	defer r.lock.Unlock()

	existing, ok := r.breakers[name]
	if ok {
		return existing
	}

	created := newCircuitBreaker(name, NewOptions())
	r.breakers[name] = created

	return created
}
