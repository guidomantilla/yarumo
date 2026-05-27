package health

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cutils "github.com/guidomantilla/yarumo/core/common/utils"
)

// health implements [Health].
//
// It stores [Check] instances in registration order and runs them
// concurrently bounded by the configured concurrency limit when
// [Health.Status] is invoked.
//
// It is safe for concurrent use by multiple goroutines.
type health struct {
	mu          sync.RWMutex
	checks      []Check
	concurrency int
}

// NewHealth creates a [Health] aggregator with the given options.
//
// Defaults are documented on [NewOptions]. The returned aggregator is safe
// for concurrent use; [Health.Status] runs the registered probes
// concurrently (bounded by the configured concurrency limit) and combines
// their [Result.Status] values using a worst-status-wins rule.
func NewHealth(opts ...Option) Health {
	options := NewOptions(opts...)

	return &health{
		checks:      nil,
		concurrency: options.concurrency,
	}
}

// Register adds a [Check] to the aggregator. Nil checks are ignored.
func (h *health) Register(check Check) {
	cassert.NotNil(h, "health is nil")

	if cutils.Nil(check) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.checks = append(h.checks, check)
}

// Status runs all registered probes concurrently (bounded by the configured
// concurrency limit) and returns the aggregated worst-status together with
// the per-check Results in registration order.
//
// When ctx is nil, an empty [StatusUnknown] result is returned without
// invoking any probe. When ctx is already cancelled before Status is called,
// no probe is invoked and the function returns immediately with
// [StatusUnknown] and an empty Results slice. When ctx is cancelled mid-way,
// already-running probes are responsible for honouring ctx and returning
// promptly; the aggregator does not interrupt them but stops scheduling new
// ones.
func (h *health) Status(ctx context.Context) (Status, []Result) {
	cassert.NotNil(h, "health is nil")

	if cutils.Nil(ctx) {
		return StatusUnknown, nil
	}

	h.mu.RLock()
	// Snapshot the checks slice so concurrent Register calls do not race
	// with the probing goroutines.
	snapshot := make([]Check, len(h.checks))
	copy(snapshot, h.checks)
	limit := h.concurrency
	h.mu.RUnlock()

	if len(snapshot) == 0 {
		return StatusUnknown, nil
	}

	results := make([]Result, len(snapshot))

	h.runProbes(ctx, snapshot, limit, results)

	return aggregate(results), results
}

// runProbes executes each check in snapshot concurrently, bounded by limit,
// and writes the outcome of each probe into results at the matching index.
// Pre-cancelled contexts cause runProbes to exit immediately without
// invoking any probe, leaving results filled with their zero value (which
// aggregates to StatusUnknown). The caller must ensure limit >= 1 and
// snapshot is non-empty — those preconditions are enforced by Status.
func (h *health) runProbes(ctx context.Context, snapshot []Check, limit int, results []Result) {
	cassert.NotNil(h, "health is nil")

	if limit > len(snapshot) {
		limit = len(snapshot)
	}

	sem := make(chan struct{}, limit)

	var wg sync.WaitGroup

	for i, check := range snapshot {
		// Stop scheduling new probes once the caller cancels ctx — already
		// scheduled probes will see the cancellation through their own ctx.
		err := ctx.Err()
		if err != nil {
			break
		}

		wg.Add(1)

		sem <- struct{}{}

		go func(idx int, c Check) {
			defer wg.Done()
			defer func() { <-sem }()

			results[idx] = probeOne(ctx, c)
		}(i, check)
	}

	wg.Wait()
}
