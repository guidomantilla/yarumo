package pollingconsumer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// pollingConsumer is the Polling Consumer endpoint implementation. It
// owns N worker goroutines (per WithMaxConcurrency) that poll the
// source PollableChannel via Receive in a loop and dispatch each pulled
// message to the user-supplied Handler. Workers exit on source close,
// ctx cancellation or Stop signal; in-flight Handler invocations
// complete before the worker returns.
type pollingConsumer[T any] struct {
	name           string
	src            messaging.PollableChannel[T]
	handler        messaging.Handler[T]
	pollInterval   time.Duration
	maxConcurrency int
	errorHandler   messaging.ErrorHandler

	started      atomic.Bool
	workerCtx    context.Context
	workerCancel context.CancelFunc
	workerWG     sync.WaitGroup

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once
}

// NewPollingConsumer constructs a Polling Consumer that polls src in a
// worker loop and dispatches each pulled Message[T] to handler. The
// consumer is not running on return; call lifecycle.Build (or Start
// directly) to spawn the worker goroutines.
//
// name is used in lifecycle logs and must be non-empty. src and handler
// are mandatory. Optional behaviors:
//
//   - WithMaxConcurrency sets the worker pool size (default 1).
//   - WithPollInterval inserts a pause between Receive calls.
//   - WithErrorHandler observes Handler errors/panics and unexpected
//     Receive errors.
func NewPollingConsumer[T any](name string, src messaging.PollableChannel[T], handler messaging.Handler[T], opts ...Option) PollingConsumer[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(handler, "handler is nil")

	options := NewOptions(opts...)

	return &pollingConsumer[T]{
		name:           name,
		src:            src,
		handler:        handler,
		pollInterval:   options.pollInterval,
		maxConcurrency: options.maxConcurrency,
		errorHandler:   options.errorHandler,
		done:           make(chan struct{}),
	}
}

// Name returns the consumer's identity used in lifecycle logs.
func (c *pollingConsumer[T]) Name() string {
	cassert.NotNil(c, "polling consumer is nil")

	return c.name
}

// Start spawns the worker goroutines and returns immediately.
// lifecycle.Component worker-style contract. Start is idempotent — a
// second invocation returns nil without spawning additional workers.
// Each worker captures workerCtx (derived from the ctx passed to Start)
// so cancelling workerCtx during Stop wakes blocked Receive calls.
func (c *pollingConsumer[T]) Start(ctx context.Context) error {
	cassert.NotNil(c, "polling consumer is nil")

	c.startOnce.Do(func() {
		workerCtx, workerCancel := context.WithCancel(ctx)
		c.workerCtx = workerCtx
		c.workerCancel = workerCancel
		c.started.Store(true)

		// Start every worker before launching awaitDrain so the waiter
		// never observes an empty WaitGroup mid-spawn. wg.Go adds 1 and
		// spawns the goroutine atomically; awaitDrain calls wg.Wait
		// only after all Add(1)s have happened, avoiding the classic
		// Wait-vs-Add race.
		for range c.maxConcurrency {
			c.workerWG.Go(func() { c.run(workerCtx) })
		}

		go c.awaitDrain()
	})

	return nil
}

// awaitDrain closes done exactly once after every worker goroutine has
// exited (post-Stop or post-source-close).
func (c *pollingConsumer[T]) awaitDrain() {
	c.workerWG.Wait()
	c.doneOnce.Do(func() { close(c.done) })
}

// Stop cancels the worker ctx (waking blocked Receive calls) and waits
// for the workers to drain up to ctx's deadline. Stop is idempotent.
// In-flight Handler invocations complete before the worker exits; Done
// closes after the last worker returns.
func (c *pollingConsumer[T]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "polling consumer is nil")

	c.stopOnce.Do(func() {
		if c.workerCancel != nil {
			c.workerCancel()
		}
	})

	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	}
}

// Done returns the channel that is closed after every worker goroutine
// has exited.
func (c *pollingConsumer[T]) Done() <-chan struct{} {
	cassert.NotNil(c, "polling consumer is nil")

	return c.done
}

// run is the per-worker loop. It calls Receive, dispatches the message
// to the user handler under panic recovery, optionally pauses for
// pollInterval and repeats until the source closes, the workerCtx
// expires or an unrecoverable Receive error surfaces.
func (c *pollingConsumer[T]) run(workerCtx context.Context) {
	for {
		select {
		case <-workerCtx.Done():
			return
		default:
		}

		msg, err := c.src.Receive(workerCtx)
		if err != nil {
			c.classifyReceiveError(workerCtx, err)

			return
		}

		c.dispatch(workerCtx, msg)

		if c.pollInterval > 0 {
			c.sleep(workerCtx)
		}
	}
}

// classifyReceiveError decides whether a Receive error is a clean
// termination signal (channel closed / ctx cancelled) or a real
// failure. Termination signals exit silently; real failures are routed
// to the ErrorHandler wrapped in ErrPollingConsumer(ErrPollFailed, ...).
func (c *pollingConsumer[T]) classifyReceiveError(ctx context.Context, err error) {
	if errors.Is(err, messaging.ErrChannelClosed) {
		return
	}

	if ctx.Err() != nil {
		return
	}

	if c.errorHandler != nil {
		c.errorHandler(ctx, nil, ErrPollingConsumer(ErrPollFailed, err))
	}
}

// dispatch invokes the user handler with panic recovery and routes
// errors/panics through the configured ErrorHandler. msg is forwarded
// to the hook unchanged so observers can inspect the failing payload.
func (c *pollingConsumer[T]) dispatch(ctx context.Context, msg messaging.Message[T]) {
	err := c.invoke(ctx, msg)
	if err == nil {
		return
	}

	if c.errorHandler == nil {
		return
	}

	c.errorHandler(ctx, msg, err)
}

// invoke runs handler with panic recovery. Returns nil on success, an
// ErrPollingConsumer(ErrHandlerFailed, err) wrapping the handler error,
// or ErrPollingConsumer(ErrHandlerPanic, ...) wrapping the recovered
// panic value.
func (c *pollingConsumer[T]) invoke(ctx context.Context, msg messaging.Message[T]) (err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		err = ErrPollingConsumer(ErrHandlerPanic, fmt.Errorf("%v", rec))
	}()

	handlerErr := c.handler(ctx, msg)
	if handlerErr != nil {
		return ErrPollingConsumer(ErrHandlerFailed, handlerErr)
	}

	return nil
}

// sleep waits pollInterval or until workerCtx cancels. The timer is
// stopped on early exit so the goroutine doesn't leak a pending timer.
func (c *pollingConsumer[T]) sleep(workerCtx context.Context) {
	timer := time.NewTimer(c.pollInterval)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-workerCtx.Done():
	}
}
