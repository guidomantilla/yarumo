package aggregator

import (
	"context"
	"fmt"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// sweeperTickDivisor sets the sweeper tick interval as a fraction of
// the configured group timeout. Smaller divisor → coarser polling →
// lower CPU but worse worst-case latency past timeout. The value of 2
// guarantees expired groups are released within at most 1.5 × timeout
// even in the worst-case scheduling.
const sweeperTickDivisor = 2

// minSweeperTick caps the minimum sweeper tick interval to avoid
// pathological CPU burn when WithGroupTimeout is set to a tiny value.
const minSweeperTick = 10 * time.Millisecond

// NewAggregator constructs an Aggregator that subscribes to src,
// collects each Message[T] into a group keyed by CorrelationFn, and
// forwards Message[U] to dst once the configured CompletionStrategy
// fires.
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// aggregate are mandatory. At least one of WithCompletionFn,
// WithCompletionSize or WithGroupTimeout MUST be configured —
// constructing an Aggregator with none is a caller bug and panics here.
//
// Optional behaviors:
//
//   - WithCorrelationFn replaces the default Headers.CorrelationID
//     extractor.
//   - WithCompletionSize, WithCompletionFn, WithGroupTimeout configure
//     completion strategies; any subset is allowed and the first one
//     that fires wins.
//   - WithMaxGroups caps in-flight groups (default 1000) for memory
//     bounding; n+1-th group fires WithErrorHandler.
//   - WithErrorHandler / WithDropHandler install observability hooks.
func NewAggregator[T, U any](name string, src messaging.Channel[T], dst messaging.Channel[U], aggregate AggregateFn[T, U], opts ...Option[T]) Aggregator[T, U] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(aggregate, "aggregate function is nil")

	options := NewOptions(opts...)

	if options.completion == nil && options.completionSize <= 0 && options.groupTimeout <= 0 {
		panic("aggregator requires at least one completion strategy: WithCompletionSize, WithCompletionFn or WithGroupTimeout")
	}

	return &aggregator[T, U]{
		name:           name,
		src:            src,
		dst:            dst,
		aggregate:      aggregate,
		correlation:    options.correlation,
		completion:     options.completion,
		completionSize: options.completionSize,
		groupTimeout:   options.groupTimeout,
		maxGroups:      options.maxGroups,
		errorHandler:   options.errorHandler,
		dropHandler:    options.dropHandler,
		done:           make(chan struct{}),
		groups:         map[string]*group[T]{},
	}
}

// Name returns the aggregator's identity used in lifecycle logs.
func (a *aggregator[T, U]) Name() string {
	cassert.NotNil(a, "aggregator is nil")

	return a.name
}

// Start registers the aggregating handler as a subscriber on the source
// channel and (when WithGroupTimeout is configured) spawns the
// background sweeper goroutine. It satisfies the lifecycle.Component
// worker-style contract: Start returns immediately after subscription
// and sweeper spawn. Start is idempotent — a second invocation returns
// nil without re-subscribing.
func (a *aggregator[T, U]) Start(ctx context.Context) error {
	cassert.NotNil(a, "aggregator is nil")

	var startErr error

	a.startOnce.Do(func() {
		cancel, err := a.src.Subscribe(a.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		a.mu.Lock()
		a.cancel = cancel
		a.mu.Unlock()

		if a.groupTimeout > 0 {
			workerCtx, workerCancel := context.WithCancel(ctx)
			a.workerCancel = workerCancel

			a.workerWG.Go(func() {
				a.runSweeper(workerCtx)
			})
		}
	})

	return startErr
}

// Stop cancels the source-channel subscription, signals the sweeper to
// exit, waits for it, drains every remaining in-flight group through
// the normal release path (so consumers see the partial aggregates),
// and closes Done. Stop is idempotent per the lifecycle.Component
// contract. It returns lifecycle.ErrShutdown wrapping
// lifecycle.ErrShutdownTimeout when ctx expires before drain completes.
func (a *aggregator[T, U]) Stop(ctx context.Context) error {
	cassert.NotNil(a, "aggregator is nil")

	a.stopOnce.Do(func() {
		a.mu.Lock()
		cancel := a.cancel
		a.cancel = nil
		a.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		if a.workerCancel != nil {
			a.workerCancel()
		}

		a.workerWG.Wait()

		a.drainRemaining(ctx)

		a.doneOnce.Do(func() { close(a.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has drained
// remaining groups.
func (a *aggregator[T, U]) Done() <-chan struct{} {
	cassert.NotNil(a, "aggregator is nil")

	return a.done
}

// handle is the Handler[T] subscribed on the source channel. It
// extracts the correlation key, appends the message to (or creates) the
// matching group, checks completion, and releases the group when
// complete. Always returns nil so aggregation concerns never propagate
// to the source channel's Send caller.
func (a *aggregator[T, U]) handle(ctx context.Context, msg messaging.Message[T]) error {
	key := a.correlation(msg)
	if key == "" {
		a.reportDrop(ctx, msg)

		return nil
	}

	a.mu.Lock()

	g, exists := a.groups[key]
	if !exists {
		if len(a.groups) >= a.maxGroups {
			a.mu.Unlock()
			a.reportError(ctx, msg, ErrAggregator(ErrMaxGroupsExceeded))

			return nil
		}

		now := time.Now()
		g = &group[T]{firstSeen: now}
		a.groups[key] = g
	}

	g.msgs = append(g.msgs, msg)
	g.lastSeen = time.Now()

	if !a.isComplete(g) {
		a.mu.Unlock()

		return nil
	}

	delete(a.groups, key)
	snapshot := g.msgs
	a.mu.Unlock()

	a.release(ctx, snapshot, nil)

	return nil
}

// isComplete returns true when any size-based or predicate-based
// completion strategy fires. Timeout-based completion is the sweeper's
// job and is intentionally NOT checked here. Caller must hold a.mu.
func (a *aggregator[T, U]) isComplete(g *group[T]) bool {
	if a.completionSize > 0 && len(g.msgs) >= a.completionSize {
		return true
	}

	if a.completion != nil && a.completion(g.msgs) {
		return true
	}

	return false
}

// runSweeper periodically scans the in-flight groups and releases any
// whose lastSeen is older than groupTimeout. The interval is half the
// configured timeout (bounded by minSweeperTick) so an expired group
// is released within at most 1.5 × timeout. The goroutine exits when
// workerCtx is cancelled by Stop.
func (a *aggregator[T, U]) runSweeper(workerCtx context.Context) {
	interval := max(a.groupTimeout/sweeperTickDivisor, minSweeperTick)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-workerCtx.Done():
			return
		case <-ticker.C:
			a.sweepExpired(workerCtx)
		}
	}
}

// sweepExpired releases every group whose lastSeen is older than
// groupTimeout. It snapshots expired keys and their message slices
// under the lock, then performs the release work (aggregate + Send +
// hook invocations) outside the lock so concurrent producers do not
// stall waiting for AggregateFn or destination Send to complete.
func (a *aggregator[T, U]) sweepExpired(ctx context.Context) {
	now := time.Now()

	a.mu.Lock()

	releases := make([][]messaging.Message[T], 0, len(a.groups))

	for key, g := range a.groups {
		if now.Sub(g.lastSeen) < a.groupTimeout {
			continue
		}

		releases = append(releases, g.msgs)
		delete(a.groups, key)
	}
	a.mu.Unlock()

	for _, msgs := range releases {
		a.release(ctx, msgs, ErrGroupExpired)
	}
}

// drainRemaining is invoked from Stop after the sweeper has exited and
// the source subscription has been cancelled. It releases every
// remaining group through the normal release path so partial groups
// land at the destination instead of being silently dropped.
func (a *aggregator[T, U]) drainRemaining(ctx context.Context) {
	a.mu.Lock()

	snapshots := make([][]messaging.Message[T], 0, len(a.groups))
	for key, g := range a.groups {
		snapshots = append(snapshots, g.msgs)
		delete(a.groups, key)
	}
	a.mu.Unlock()

	for _, msgs := range snapshots {
		a.release(ctx, msgs, nil)
	}
}

// release folds the group into a single Message[U] via AggregateFn
// (under panic recovery) and forwards it to dst. Failures route through
// the ErrorHandler; reason is joined into the error when non-nil so
// callers can distinguish timeout-released aggregations from normal
// completions (reason == ErrGroupExpired).
func (a *aggregator[T, U]) release(ctx context.Context, msgs []messaging.Message[T], reason error) {
	if len(msgs) == 0 {
		return
	}

	out, err := a.aggregateWithRecover(msgs)
	if err != nil {
		if reason != nil {
			a.reportError(ctx, nil, ErrAggregator(reason, err))

			return
		}

		a.reportError(ctx, nil, err)

		return
	}

	err = a.dst.Send(ctx, out)
	if err != nil {
		causes := []error{ErrForwardFailed, err}
		if reason != nil {
			causes = append(causes, reason)
		}

		a.reportError(ctx, nil, ErrAggregator(causes...))
	}
}

// aggregateWithRecover invokes AggregateFn under panic recovery. A
// returned error becomes ErrAggregator(ErrAggregateFnFailed, err); a
// panic becomes ErrAggregator(ErrAggregateFnFailed, "panic: <value>").
func (a *aggregator[T, U]) aggregateWithRecover(msgs []messaging.Message[T]) (out messaging.Message[U], err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		err = ErrAggregator(ErrAggregateFnFailed, fmt.Errorf("panic: %v", rec))
	}()

	out, err = a.aggregate(msgs)
	if err != nil {
		return out, ErrAggregator(ErrAggregateFnFailed, err)
	}

	return out, nil
}

// reportError forwards err to the configured ErrorHandler. The handler
// is guaranteed non-nil by NewOptions; the nil-guard is defensive only.
func (a *aggregator[T, U]) reportError(ctx context.Context, msg any, err error) {
	if a.errorHandler == nil {
		return
	}

	a.errorHandler(ctx, msg, err)
}

// reportDrop forwards msg to the configured DropHandler. DropHandler is
// nil by default (silent drops); the guard skips invocation in that
// case.
func (a *aggregator[T, U]) reportDrop(ctx context.Context, msg any) {
	if a.dropHandler == nil {
		return
	}

	a.dropHandler(ctx, msg)
}
