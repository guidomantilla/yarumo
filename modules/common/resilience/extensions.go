package resilience

// Package-level default registries. Both are lazy and goroutine-free; callers
// can either use these defaults or construct their own via
// NewCircuitBreakerRegistry / NewRateLimiterRegistry.
var (
	// DefaultCircuitBreakerRegistry is the package-level default registry of
	// circuit breakers.
	DefaultCircuitBreakerRegistry = NewCircuitBreakerRegistry()
	// DefaultRateLimiterRegistry is the package-level default registry of
	// rate limiters.
	DefaultRateLimiterRegistry = NewRateLimiterRegistry()
)
