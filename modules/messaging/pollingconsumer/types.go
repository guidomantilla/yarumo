// Package pollingconsumer provides a Polling Consumer endpoint over
// messaging.PollableChannel[T].
//
// A Polling Consumer is the pull-based counterpart of the event-driven
// consumer (Channel[T].Subscribe): instead of registering a Handler and
// waiting for the channel to push messages, the consumer spawns a
// worker goroutine that calls Receive in a loop and dispatches each
// pulled message to a user-supplied Handler. The pattern decouples
// downstream readiness from upstream production — the consumer
// dequeues only when it is ready to process, and the PollableChannel
// applies natural backpressure on the producer when the consumer is
// slow (Send blocks once the buffer fills).
//
// # Worker pool
//
// WithMaxConcurrency selects how many worker goroutines compete to
// poll the same source. With the default 1, the consumer is strictly
// sequential — at most one message is in flight. With N > 1, N workers
// poll the channel concurrently and the Handler must be safe for
// concurrent invocation (the Handler is the user's "service activator"
// — its concurrency requirements are the user's problem).
//
// WithPollInterval lets the worker pause between Receive calls. The
// default 0 polls immediately on each loop iteration; the
// PollableChannel's blocking Receive already provides natural
// backpressure (the call blocks until a message is available, ctx
// expires, or the channel closes), so an explicit interval is rarely
// needed and exists mainly for resource-shaping when callers want to
// rate-limit calls into an external poller.
//
// # Lifecycle
//
// PollingConsumer implements common/lifecycle.Component (worker-style):
// Start spawns the worker goroutines and returns immediately; Stop
// signals the workers to exit and waits up to ctx for them to drain.
// Workers exit cleanly when:
//
//   - the source PollableChannel is closed (Receive returns
//     ErrChannelClosed once the buffer is drained);
//   - the worker's ctx is cancelled (Receive returns ErrTimeout
//     wrapping the ctx error);
//   - Stop signals the internal stop channel.
//
// In-flight Handler invocations complete before the worker exits; Done
// closes after the last worker returns.
//
// # Error handling
//
// Handler errors and Handler panics are routed to WithErrorHandler
// (defaulting to messaging.DefaultErrorHandler which logs via
// common/log). Receive errors that signal a clean termination
// (ErrChannelClosed, ctx-cancellation) are NOT reported through the
// hook — they're expected control flow. Any other Receive error is
// wrapped in ErrPollingConsumer(ErrPollFailed, ...) and reported, and
// then the worker exits (the channel is in an indeterminate state).
//
// Unlike push-based EIP patterns there is no DropHandler — polling
// consumers do not gate messages; everything dequeued reaches the
// Handler.
package pollingconsumer

import (
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ PollingConsumer[any] = (*pollingConsumer[any])(nil)

	_ ErrPollingConsumerFn = ErrPollingConsumer
)

// PollingConsumer is the public interface for a Polling Consumer
// endpoint. It embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface preserves
// "this is a PollingConsumer" semantics and the type stays open to
// future polling-consumer-specific methods without breaking callers.
type PollingConsumer[T any] interface {
	lifecycle.Component
}

// ErrPollingConsumerFn is the function type for ErrPollingConsumer.
type ErrPollingConsumerFn func(causes ...error) error
