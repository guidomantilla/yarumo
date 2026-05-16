package resilience

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// DefaultCircuitBreakerRegistry is the package-level default registry of
// circuit breakers. It is lazy and goroutine-free; callers can either use this
// default or construct their own via NewCircuitBreakerRegistry.
var DefaultCircuitBreakerRegistry = NewCircuitBreakerRegistry()

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
