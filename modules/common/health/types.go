// Package health provides primitive types and a synchronous aggregator for
// health-check probes.
//
// Scope: this package is intentionally limited to **types + a synchronous
// aggregator**. It contains no HTTP endpoint, no periodic background
// probing, and no lifecycle integration — those concerns live in
// `modules/health/`.
//
// Aggregation contract: when [Health.Status] is invoked, it runs all
// registered [Check.Probe] calls concurrently (bounded by the configured
// concurrency limit) and aggregates the individual [Result.Status] values
// using a **worst-status-wins** rule. The ordering, from worst to best, is:
// [StatusUnhealthy] > [StatusDegraded] > [StatusHealthy] > [StatusUnknown].
// When no checks are registered, the aggregated status is [StatusUnknown].
//
// Concurrency: [Health] implementations must be safe for concurrent use by
// multiple goroutines. [Check] implementations should also be safe for
// concurrent use because [Health.Status] may invoke [Check.Probe] from any
// worker goroutine.
package health

import (
	"context"
	"time"
)

var (
	_ Health = (*health)(nil)

	_ AggregateFn = aggregate
)

// AggregateFn is the function type for the worst-status-wins aggregation.
type AggregateFn func(results []Result) Status

// Status represents the health classification of a probe or an aggregate.
//
// Ordering, from worst to best:
// [StatusUnhealthy] > [StatusDegraded] > [StatusHealthy] > [StatusUnknown].
type Status int

// Status values, ordered so that a higher integer means a worse status.
// This ordering is what enables the worst-status-wins aggregation rule.
const (
	// StatusUnknown indicates the probe has no information yet or no checks are registered.
	StatusUnknown Status = iota
	// StatusHealthy indicates the probe passed without issues.
	StatusHealthy
	// StatusDegraded indicates the probe completed with non-fatal warnings.
	StatusDegraded
	// StatusUnhealthy indicates the probe failed.
	StatusUnhealthy
)

// String returns the textual representation of the status.
func (s Status) String() string {
	switch s {
	case StatusUnknown:
		return "unknown"
	case StatusHealthy:
		return "healthy"
	case StatusDegraded:
		return "degraded"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// Result is the outcome of a single [Check.Probe] invocation.
type Result struct {
	// Name is the name of the [Check] that produced this Result.
	Name string
	// Status is the health classification produced by the probe.
	Status Status
	// Message is a human-readable description of the outcome (optional).
	Message string
	// Details carries optional structured metadata produced by the probe.
	Details map[string]any
	// Duration is the wall-clock time taken by the probe to produce this Result.
	Duration time.Duration
}

// Check defines a single health probe.
//
// Implementations must be safe for concurrent use because [Health.Status]
// may invoke [Check.Probe] from any worker goroutine. Implementations should
// honour ctx cancellation and return promptly when ctx is Done.
type Check interface {
	// Name returns the unique identifier of the check.
	Name() string
	// Probe performs the health check and returns its [Result].
	Probe(ctx context.Context) Result
}

// Health aggregates many [Check] instances into a single status.
//
// Implementations must be safe for concurrent use. [Health.Status] runs the
// registered probes concurrently (bounded by the configured concurrency
// limit) and combines their [Result.Status] values using a worst-status-wins
// rule. See the package documentation for the full contract.
type Health interface {
	// Register adds a [Check] to the aggregator. Subsequent calls to
	// [Health.Status] will include this check.
	Register(check Check)
	// Status runs all registered probes concurrently (bounded) and returns
	// the aggregated worst-status together with the per-check Results in
	// registration order.
	Status(ctx context.Context) (Status, []Result)
}
