package resilience

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
)

type CircuitBreakerRegistry struct {
	mu       sync.Mutex
	breakers map[string]*gobreaker.CircuitBreaker
}

func NewCircuitBreakerRegistry() *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
	}
}

func (registry *CircuitBreakerRegistry) Get(name string) *gobreaker.CircuitBreaker {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	cb, ok := registry.breakers[name]
	if ok {
		return cb
	}

	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Warn().Str("stage", "runtime").Str("component", name).Str("from", from.String()).Str("to", to.String()).Msg("circuit breaker state changed")
		},
	}
	cb = gobreaker.NewCircuitBreaker(settings)
	registry.breakers[name] = cb
	return cb
}
