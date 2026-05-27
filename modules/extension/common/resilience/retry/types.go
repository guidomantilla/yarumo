// Package retry provides a generic retry implementation backed by
// github.com/avast/retry-go/v4. It implements the Retry contract
// defined in modules/core/common/resilience/retry/.
//
// Construct an instance via NewRetry(opts ...Option). Each call returns an
// independent Retry; there is no registry, no global singleton, and no
// pluggable function fields. Callers that need multiple policies construct
// multiple instances.
//
// Concurrency: every method on the returned Retry is safe for concurrent
// use by multiple goroutines.
//
// Goroutine-free: NewRetry does not spawn goroutines. Each Do call runs
// the retry loop on the caller's goroutine; delays between attempts are
// time.Sleep equivalents bounded by ctx.
package retry

import (
	cretry "github.com/guidomantilla/yarumo/core/common/resilience/retry"
)

var (
	_ cretry.Retry = (*retry)(nil)
)
