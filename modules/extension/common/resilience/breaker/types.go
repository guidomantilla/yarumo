// Package breaker provides a circuit breaker implementation backed by
// github.com/sony/gobreaker. It implements the Breaker contract defined
// in modules/core/common/resilience/breaker/.
//
// Construct an instance via NewBreaker(opts ...Option). Each call returns
// an independent Breaker; there is no registry, no global singleton, and
// no pluggable function fields. Callers that need multiple breakers
// construct multiple instances.
//
// Concurrency: every method on the returned Breaker is safe for
// concurrent use by multiple goroutines.
//
// Goroutine-free: NewBreaker does not spawn goroutines. State transitions
// happen synchronously, observed on the next Execute call.
package breaker

import (
	cbreaker "github.com/guidomantilla/yarumo/core/common/resilience/breaker"
)

var (
	_ cbreaker.Breaker = (*breaker)(nil)
)
