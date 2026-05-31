package scattergather

import (
	"context"
	"errors"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
	"github.com/guidomantilla/yarumo/messaging/patterns/routers/aggregator"
	"github.com/guidomantilla/yarumo/messaging/patterns/routers/recipientlist"
)

// orphanSweepDivisor sets the orphan-sweeper tick interval as a
// fraction of the configured group timeout. The orphan sweeper evicts
// expected-map entries whose workers never replied (so the internal
// Aggregator never sees any group for them and never times them out).
// The value of 2 guarantees an orphan entry is evicted within at most
// 1.5 × timeout even in the worst-case scheduling.
const orphanSweepDivisor = 2

// minOrphanSweepTick caps the minimum orphan-sweeper tick interval to
// avoid pathological CPU burn when WithGroupTimeout is set to a tiny
// value.
const minOrphanSweepTick = 10 * time.Millisecond

// orphanTTLMultiplier multiplies WithGroupTimeout to compute the TTL
// of an expected-map entry that has not yet been observed by the
// internal Aggregator (i.e. no replies arrived for it). Set higher
// than 1× so the natural Aggregator timeout path wins when at least
// one reply did arrive — only true orphans (no replies at all) are
// evicted by this sweeper.
const orphanTTLMultiplier = 2

// errPartialDrop is the internal sentinel returned by the wrapped
// AggregateFn when a gather is released by the Aggregator's sweeper
// without reaching its expected reply count (a worker never replied).
// We detect it in the wrapped Aggregator ErrorHandler and route the
// drop to the user-facing DropHandler instead of the ErrorHandler.
// This sentinel never escapes the package.
var errPartialDrop = errors.New("scattergather partial gather dropped on timeout")

// NewScatterGather constructs a Scatter-Gather pattern wiring an
// internal Recipient List (the scatter half) and an internal
// Aggregator (the gather half).
//
// Arguments:
//
//   - name is used in lifecycle logs and must be non-empty.
//   - src is the source Channel[T] from which requests arrive.
//   - workers maps every selector key to a worker destination
//     Channel[T] that receives the scattered request.
//   - replyChan is the Channel[T] every worker publishes its reply on
//     (with the SAME CorrelationID as the request).
//   - aggregateDst is the final Channel[U] the gathered Message[U]
//     lands on once all expected replies arrive.
//   - selector returns the worker keys for each request; an empty
//     slice fires the DropHandler with no scatter.
//   - aggregate folds the collected replies into the final
//     Message[U]; a non-nil error routes through ErrorHandler.
//
// Options:
//
//   - WithGroupTimeout is REQUIRED — without it a worker that never
//     replies would stall the gather forever. Constructing a
//     ScatterGather with no group timeout is a caller bug and panics.
//   - WithMaxConcurrentScatters caps in-flight gathers
//     (default 1000); exceeding it fires WithErrorHandler with
//     ErrMaxScattersExceeded.
//   - WithErrorHandler / WithDropHandler install observability hooks.
func NewScatterGather[T, U any](name string, src messaging.Channel[T], workers map[string]messaging.Channel[T], replyChan messaging.Channel[T], aggregateDst messaging.Channel[U], selector SelectorFn[T], aggregate AggregateFn[T, U], opts ...Option[T]) ScatterGather[T, U] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(workers, "workers map is nil")
	cassert.NotNil(replyChan, "reply channel is nil")
	cassert.NotNil(aggregateDst, "aggregate destination channel is nil")
	cassert.NotNil(selector, "selector function is nil")
	cassert.NotNil(aggregate, "aggregate function is nil")

	if len(workers) == 0 {
		panic("scattergather requires at least one worker channel")
	}

	options := NewOptions(opts...)

	if options.groupTimeout <= 0 {
		panic("scattergather requires WithGroupTimeout to bound partial-gather lifetime")
	}

	sg := &scatterGather[T, U]{
		name:                  name,
		src:                   src,
		workers:               workers,
		replyChan:             replyChan,
		aggregateDst:          aggregateDst,
		selector:              selector,
		aggregate:             aggregate,
		groupTimeout:          options.groupTimeout,
		maxConcurrentScatters: options.maxConcurrentScatters,
		errorHandler:          options.errorHandler,
		dropHandler:           options.dropHandler,
		done:                  make(chan struct{}),
		expected:              map[string]expectation{},
	}

	sg.scatterer = recipientlist.NewRecipientList(name+"-scatter", src, sg.wrappedSelector, workers,
		recipientlist.WithErrorHandler(sg.recipientListError),
		recipientlist.WithDropHandler(sg.recipientListDrop))

	sg.gatherer = aggregator.NewAggregator(name+"-gather", replyChan, aggregateDst, sg.wrappedAggregate,
		aggregator.WithCompletionFn(sg.completion),
		aggregator.WithGroupTimeout[T](options.groupTimeout),
		aggregator.WithErrorHandler[T](sg.aggregatorError),
		aggregator.WithDropHandler[T](sg.aggregatorDrop))

	return sg
}

// Name returns the scatter-gather's identity used in lifecycle logs.
func (s *scatterGather[T, U]) Name() string {
	cassert.NotNil(s, "scatter-gather is nil")

	return s.name
}

// Start spawns the internal Aggregator first (so it is ready to
// receive worker replies before any scatter happens), then the
// internal Recipient List, and finally the orphan sweeper goroutine
// that evicts expected-map entries whose workers never replied. Start
// is idempotent — a second invocation returns nil without
// re-subscribing.
func (s *scatterGather[T, U]) Start(ctx context.Context) error {
	cassert.NotNil(s, "scatter-gather is nil")

	var startErr error

	s.startOnce.Do(func() {
		err := s.gatherer.Start(ctx)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		err = s.scatterer.Start(ctx)
		if err != nil {
			_ = s.gatherer.Stop(ctx)
			startErr = lifecycle.ErrStart(err)

			return
		}

		workerCtx, workerCancel := context.WithCancel(ctx)
		s.workerCancel = workerCancel

		s.workerWG.Go(func() {
			s.runOrphanSweeper(workerCtx)
		})
	})

	return startErr
}

// Stop reverses the start order: it stops the internal Recipient List
// first (no new requests scatter), then the internal Aggregator
// (which drains in-flight gathers via the standard release path —
// timed-out partial gathers route to WithDropHandler via the wrapped
// AggregateFn). Stop is idempotent per the lifecycle.Component
// contract. It returns lifecycle.ErrShutdown wrapping
// lifecycle.ErrShutdownTimeout when ctx expires before both halves
// have stopped.
func (s *scatterGather[T, U]) Stop(ctx context.Context) error {
	cassert.NotNil(s, "scatter-gather is nil")

	s.stopOnce.Do(func() {
		_ = s.scatterer.Stop(ctx)

		if s.workerCancel != nil {
			s.workerCancel()
		}

		s.workerWG.Wait()

		_ = s.gatherer.Stop(ctx)

		s.doneOnce.Do(func() { close(s.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has drained both
// internal halves.
func (s *scatterGather[T, U]) Done() <-chan struct{} {
	cassert.NotNil(s, "scatter-gather is nil")

	return s.done
}

// wrappedSelector intercepts the user-supplied SelectorFn so it can
// (1) record the expected reply count per correlation id at scatter
// time and (2) enforce the WithMaxConcurrentScatters cap. The
// returned keys are passed straight through to the internal Recipient
// List for fan-out. Empty results pass through to the Recipient
// List's drop handler unchanged; user-selector errors and the
// MaxScatters cap surface through the Recipient List's error handler.
func (s *scatterGather[T, U]) wrappedSelector(ctx context.Context, msg messaging.Message[T]) ([]string, error) {
	keys, err := s.selector(ctx, msg)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	corrID := msg.Headers.CorrelationID

	s.mu.Lock()

	if len(s.expected) >= s.maxConcurrentScatters {
		s.mu.Unlock()

		return nil, ErrMaxScattersExceeded
	}

	s.expected[corrID] = expectation{count: len(keys), scatteredAt: time.Now()}
	s.mu.Unlock()

	return keys, nil
}

// completion is the CompletionFn installed on the internal Aggregator.
// It returns true when the group has received at least the expected
// number of replies for its correlation id. The function runs under
// the Aggregator's lock; it must be cheap, which it is (one map
// lookup under our own lock).
func (s *scatterGather[T, U]) completion(group []messaging.Message[T]) bool {
	if len(group) == 0 {
		return false
	}

	corrID := group[0].Headers.CorrelationID

	s.mu.Lock()
	exp := s.expected[corrID]
	s.mu.Unlock()

	return exp.count > 0 && len(group) >= exp.count
}

// wrappedAggregate intercepts the user-supplied AggregateFn so it can
// (1) clean up the per-correlation expected-size entry at release
// time and (2) detect timeout-driven partial gathers (released by the
// Aggregator's sweeper before all expected replies arrived) and
// route them to the DropHandler via the errPartialDrop sentinel
// instead of forwarding an incomplete aggregate.
func (s *scatterGather[T, U]) wrappedAggregate(group []messaging.Message[T]) (messaging.Message[U], error) {
	if len(group) == 0 {
		return messaging.Message[U]{}, errPartialDrop
	}

	corrID := group[0].Headers.CorrelationID

	s.mu.Lock()
	exp := s.expected[corrID]
	delete(s.expected, corrID)
	s.mu.Unlock()

	if exp.count == 0 || len(group) < exp.count {
		return messaging.Message[U]{}, errPartialDrop
	}

	return s.aggregate(group)
}

// recipientListError is the ErrorHandler installed on the internal
// Recipient List. It detects ErrMaxScattersExceeded (returned from
// wrappedSelector) and reports it under our domain wrap; everything
// else is treated as a scatter failure (per-recipient missing key,
// forward Send failed, user selector error or panic).
func (s *scatterGather[T, U]) recipientListError(ctx context.Context, msg any, err error) {
	if s.errorHandler == nil {
		return
	}

	if errors.Is(err, ErrMaxScattersExceeded) {
		s.errorHandler(ctx, msg, ErrScatterGather(ErrMaxScattersExceeded))

		return
	}

	s.errorHandler(ctx, msg, ErrScatterGather(ErrScatterFailed, err))
}

// recipientListDrop is the DropHandler installed on the internal
// Recipient List. It fires when the user selector returns an empty
// slice and forwards the drop to the user-facing DropHandler.
func (s *scatterGather[T, U]) recipientListDrop(ctx context.Context, msg any) {
	if s.dropHandler == nil {
		return
	}

	s.dropHandler(ctx, msg)
}

// aggregatorError is the ErrorHandler installed on the internal
// Aggregator. It detects errPartialDrop (returned by wrappedAggregate
// on timeout-driven partial release) and routes that case to the
// user-facing DropHandler; every other failure wraps under
// ErrGatherFailed and goes to the user-facing ErrorHandler.
func (s *scatterGather[T, U]) aggregatorError(ctx context.Context, msg any, err error) {
	if errors.Is(err, errPartialDrop) {
		if s.dropHandler == nil {
			return
		}

		s.dropHandler(ctx, msg)

		return
	}

	if s.errorHandler == nil {
		return
	}

	s.errorHandler(ctx, msg, ErrScatterGather(ErrGatherFailed, err))
}

// aggregatorDrop is the DropHandler installed on the internal
// Aggregator. It fires when a reply arrives with an empty
// CorrelationID (the Aggregator's default CorrelationFn skips
// aggregation in that case) and forwards the drop to the user-facing
// DropHandler.
func (s *scatterGather[T, U]) aggregatorDrop(ctx context.Context, msg any) {
	if s.dropHandler == nil {
		return
	}

	s.dropHandler(ctx, msg)
}

// runOrphanSweeper periodically evicts expected-map entries whose
// workers never replied at all (the internal Aggregator never sees a
// group for them and never times them out). The TTL is set to
// orphanTTLMultiplier × groupTimeout so the natural Aggregator
// timeout path wins when at least one reply did arrive. The
// goroutine exits when workerCtx is cancelled by Stop.
func (s *scatterGather[T, U]) runOrphanSweeper(workerCtx context.Context) {
	interval := max(s.groupTimeout/orphanSweepDivisor, minOrphanSweepTick)
	ttl := s.groupTimeout * orphanTTLMultiplier

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-workerCtx.Done():
			return
		case <-ticker.C:
			s.sweepOrphans(workerCtx, ttl)
		}
	}
}

// sweepOrphans removes expected-map entries older than ttl and fires
// the DropHandler once per evicted orphan so observers can audit
// scatter requests that never received any worker reply. The map
// snapshot is taken under the lock, hook invocation happens outside.
func (s *scatterGather[T, U]) sweepOrphans(ctx context.Context, ttl time.Duration) {
	now := time.Now()

	s.mu.Lock()

	orphans := make([]string, 0)

	for corrID, exp := range s.expected {
		if now.Sub(exp.scatteredAt) < ttl {
			continue
		}

		orphans = append(orphans, corrID)
		delete(s.expected, corrID)
	}
	s.mu.Unlock()

	if s.dropHandler == nil {
		return
	}

	for _, corrID := range orphans {
		s.dropHandler(ctx, corrID)
	}
}
