package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// TopicChannel is the buffered async Channel[T] implementation.
//
// TopicChannel implements both Channel[T] and lifecycle.Component
// (worker-style) and uses a per-subscriber queue model: every Subscribe
// allocates its own bounded inbox and spawns a dedicated worker
// goroutine. Send fans the message out by enqueueing it on EVERY
// subscriber's inbox using the configured OverflowPolicy per inbox.
//
// The per-subscriber model gives full isolation: a slow handler only
// fills its own inbox; fast handlers keep receiving messages at line
// rate. Backpressure (when OverflowBlock is selected) applies per
// subscriber as well — a saturated slow subscriber does not block
// publishers from reaching the fast ones.
//
// TopicChannel is safe for concurrent use. Send's behavior when an
// inbox is full depends on the configured OverflowPolicy (see
// WithOverflowPolicy): Block waits per-inbox until a slot opens or
// ctx expires; DropNewest/DropOldest drop and fire the hook;
// OverflowReject (the default) returns ErrBufferFull immediately for
// that inbox. Per-inbox errors are joined via errors.Join in the
// Send return value.
//
// # Lifecycle
//
// Subscribe before Start registers the subscriber but defers spawning
// its worker; Start spawns all deferred workers (and from then on
// Subscribe spawns workers immediately). Stop closes every
// subscriber's inbox and waits for all workers to drain, bounded by
// the configured drain timeout or by ctx, whichever is tighter. Done
// closes after every worker has exited.
//
// # Cancel semantics
//
// The Cancel returned by Subscribe is idempotent and fire-and-forget:
// it removes the subscriber from the dispatch map and signals its
// worker to exit. Messages that were already buffered in the
// subscriber's inbox at the moment of Cancel may be processed before
// the worker exits, but no new messages flow to it after Cancel
// returns. In-flight Send calls that snapshotted the subscriber set
// before Cancel may still enqueue one message to the cancelled
// subscriber's inbox; that message is dropped (never processed) when
// the worker observes the done signal.
//
// # Context propagation
//
// The publisher's Send ctx travels alongside the message in the
// inbox. The handler receives a ctx whose Done / Deadline / Err
// follow the worker's lifecycle ctx (the one passed to Start) and
// whose Value lookups fall through to the Send ctx so trace span,
// correlation id and slogctx attributes propagate from publisher to
// handler. See mergeContexts in context.go.
type topic[T any] struct {
	name           string
	bufferSize     int
	drainTimeout   time.Duration
	errorHandler   ErrorHandler
	overflowPolicy OverflowPolicy

	started      atomic.Bool
	closed       atomic.Bool
	workerCtx    context.Context
	workerCancel context.CancelFunc

	workerWG sync.WaitGroup
	done     chan struct{}

	mu     sync.RWMutex
	nextID uint64
	order  []uint64
	subs   map[uint64]*subscriber[T]

	startOnce    sync.Once
	stopOnce     sync.Once
	doneOnce     sync.Once
	sentinelOnce sync.Once
}

// subscriber holds the per-Subscribe state for a TopicChannel: the
// handler itself, its private bounded inbox, and a done channel the
// worker watches for Cancel/Stop signals. The inbox is owned by this
// subscriber; only its worker drains it.
type subscriber[T any] struct {
	handler Handler[T]
	inbox   chan envelope[T]
	done    chan struct{}
}

// NewTopicChannel constructs a TopicChannel[T] with the given name and
// options. name is used in lifecycle logs and must be non-empty.
//
// The returned channel is not running; call lifecycle.Build (or Start
// directly) to spawn worker goroutines for any subscribers registered
// before Start. Subscribers registered after Start get their worker
// spawned immediately.
func NewTopicChannel[T any](name string, opts ...Option) Channel[T] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &topic[T]{
		name:           name,
		bufferSize:     options.bufferSize,
		drainTimeout:   options.drainTimeout,
		errorHandler:   options.errorHandler,
		overflowPolicy: options.overflowPolicy,
		done:           make(chan struct{}),
		subs:           map[uint64]*subscriber[T]{},
	}
}

// Name returns the channel's identity used in lifecycle logs.
func (c *topic[T]) Name() string {
	cassert.NotNil(c, "TopicChannel is nil")

	return c.name
}

// Start activates the topic. It captures ctx as the worker context
// (used for Value propagation from publisher to handler) and spawns
// the worker goroutines for any subscribers that were registered
// before Start. From now on, Subscribe will spawn workers
// immediately. Start is idempotent and Done is closed after Stop has
// drained all workers.
func (c *topic[T]) Start(ctx context.Context) error {
	cassert.NotNil(c, "TopicChannel is nil")

	c.startOnce.Do(func() {
		workerCtx, workerCancel := context.WithCancel(ctx)

		// Capture and publish workerCtx + the deferred-subscriber
		// snapshot under c.mu so Subscribe sees consistent state when
		// deciding whether to spawn its worker. Setting started=true
		// inside the lock guarantees that any Subscribe acquiring c.mu
		// after Start observes started AND reads the same workerCtx.
		c.mu.Lock()
		c.workerCtx = workerCtx
		c.workerCancel = workerCancel

		pending := make([]*subscriber[T], 0, len(c.order))
		for _, id := range c.order {
			pending = append(pending, c.subs[id])
		}

		c.started.Store(true)
		c.mu.Unlock()

		// Sentinel Add(1) keeps workerWG strictly positive between Start
		// and Stop so post-Start Subscribe calls can Add(1) without
		// racing the Wait() in awaitDrain (sync.WaitGroup forbids
		// Add-after-zero racing with Wait). The sentinel is released
		// via sentinelOnce — by Stop on the happy path, or by Start
		// itself if Stop already ran (closed observed below).
		c.workerWG.Add(1)

		for _, sub := range pending {
			c.spawnSubWorker(workerCtx, sub)
		}

		go c.awaitDrain()

		// If Stop ran before this Start completed, release the sentinel
		// ourselves so awaitDrain can converge.
		if c.closed.Load() {
			c.sentinelOnce.Do(func() { c.workerWG.Done() })
		}
	})

	return nil
}

// awaitDrain closes done exactly once after every subscriber worker
// has exited (post-Stop).
func (c *topic[T]) awaitDrain() {
	c.workerWG.Wait()
	c.doneOnce.Do(func() { close(c.done) })
}

// Stop closes every subscriber's inbox, cancels the worker context
// and waits for all workers to drain pending messages, bounded by the
// configured drain timeout or by ctx (whichever is tighter). Returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when
// the drain does not complete in time. Stop is idempotent per the
// lifecycle.Component contract.
func (c *topic[T]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "TopicChannel is nil")

	c.stopOnce.Do(func() {
		c.closed.Store(true)

		c.mu.Lock()
		cancelFn := c.workerCancel
		startedNow := c.started.Load()
		for _, sub := range c.subs {
			close(sub.inbox)
		}
		c.mu.Unlock()

		if cancelFn != nil {
			cancelFn()
		}

		if startedNow {
			// Release the sentinel installed at Start. The sentinelOnce
			// guarantees the matching Done fires exactly once, even if
			// Start runs after Stop and tries to release it itself.
			c.sentinelOnce.Do(func() { c.workerWG.Done() })
		}
	})

	waitCtx, cancel := context.WithTimeout(ctx, c.drainTimeout)
	defer cancel()

	select {
	case <-c.done:
		return nil
	case <-waitCtx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, waitCtx.Err())
	}
}

// Done returns the channel that is closed after every subscriber's
// worker has drained its inbox and exited (post-Stop).
func (c *topic[T]) Done() <-chan struct{} {
	cassert.NotNil(c, "TopicChannel is nil")

	return c.done
}

// Send fans msg out to every subscriber's inbox using the configured
// OverflowPolicy per inbox. Send returns ErrSend(ErrClosed) after
// Stop. Per-inbox errors (e.g. ErrBufferFull under OverflowReject,
// ErrTimeout under OverflowBlock with ctx expiry) are aggregated via
// errors.Join in the return value; a partial fan-out where some
// inboxes accepted the message and others did not still returns the
// joined error so callers can surface it.
func (c *topic[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(c, "TopicChannel is nil")

	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	if c.closed.Load() {
		return ErrSend(ErrClosed)
	}

	c.mu.RLock()
	snapshot := make([]*subscriber[T], 0, len(c.order))
	for _, id := range c.order {
		snapshot = append(snapshot, c.subs[id])
	}
	c.mu.RUnlock()

	if len(snapshot) == 0 {
		return nil
	}

	var errs []error
	for _, sub := range snapshot {
		select {
		case <-sub.done:
			continue
		default:
		}

		err := sendWithPolicy(ctx, sub.inbox, msg, c.overflowPolicy, c.errorHandler)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Subscribe registers handler with its own bounded inbox and worker
// goroutine. Subscribers receive messages in Subscribe order during
// the fan-out loop in Send. Cancel is idempotent. Subscribe returns
// ErrSubscribe(ErrHandlerNil) when handler is nil and
// ErrSubscribe(ErrClosed) when the channel has been stopped.
//
// If Start has not been called yet, the worker is deferred and will
// be spawned when Start runs. After Start, Subscribe spawns the
// worker immediately.
func (c *topic[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "TopicChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	if c.closed.Load() {
		return nil, ErrSubscribe(ErrClosed)
	}

	sub := &subscriber[T]{
		handler: handler,
		inbox:   make(chan envelope[T], c.bufferSize),
		done:    make(chan struct{}),
	}

	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.subs[id] = sub
	c.order = append(c.order, id)
	shouldSpawn := c.started.Load()
	workerCtx := c.workerCtx
	c.mu.Unlock()

	if shouldSpawn {
		c.spawnSubWorker(workerCtx, sub)
	}

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			c.mu.Lock()
			delete(c.subs, id)

			for i, candidate := range c.order {
				if candidate != id {
					continue
				}

				c.order = append(c.order[:i], c.order[i+1:]...)

				break
			}
			c.mu.Unlock()

			close(sub.done)
		})
	}

	return cancel, nil
}

// spawnSubWorker starts the worker goroutine for sub. The worker
// drains sub.inbox, invokes the handler under panic recovery, and
// exits when (a) sub.inbox is closed by Stop or (b) sub.done is
// closed by Cancel. Errors and recovered panics are routed through
// the configured ErrorHandler. workerCtx is captured as a parameter
// (not read from c.workerCtx) so spawnSubWorker callers can pass it
// under the same critical section that observed started=true,
// avoiding a race with Start's write.
func (c *topic[T]) spawnSubWorker(workerCtx context.Context, sub *subscriber[T]) {
	c.workerWG.Go(func() {
		for {
			select {
			case env, ok := <-sub.inbox:
				if !ok {
					return
				}

				handlerCtx := mergeContexts(workerCtx, env.sendCtx)

				err := invokeHandler(handlerCtx, env.msg, sub.handler)
				if err != nil && c.errorHandler != nil {
					c.errorHandler(handlerCtx, env.msg, err)
				}
			case <-sub.done:
				return
			}
		}
	})
}
