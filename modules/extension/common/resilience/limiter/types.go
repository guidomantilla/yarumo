// Package limiter provides a token-bucket rate limiter implementation
// backed by golang.org/x/time/rate. It implements the Limiter contract
// defined in modules/core/common/resilience/limiter/.
//
// Construct an instance via NewLimiter(opts ...Option). Each call returns
// an independent Limiter; there is no registry, no global singleton, and
// no pluggable function fields. Callers that need multiple limiters
// construct multiple instances.
//
// Concurrency: every method on the returned Limiter is safe for
// concurrent use by multiple goroutines.
//
// Goroutine-free: NewLimiter does not spawn goroutines. Rate replenishment
// happens lazily on each Allow/Wait call.
package limiter

import (
	climiter "github.com/guidomantilla/yarumo/core/common/resilience/limiter"
)

var (
	_ climiter.Limiter = (*limiter)(nil)
)
