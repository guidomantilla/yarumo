package messaging

import (
	"context"
	"sync"
	"sync/atomic"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// pollable implements PollableChannel[T] as a bounded buffered channel
// that producers feed via Send and consumers drain via Receive. Unlike
// the push-based Channel[T] implementations, pollable does not own a
// dispatcher: a message stays in the buffer until a caller explicitly
// asks for it. The struct is a thin wrapper around a Go channel plus
// a closed-state flag so Send/Receive can return the canonical domain
// errors after Close.
//
// pollable is safe for concurrent use. Multiple producers may call
// Send and multiple consumers may call Receive simultaneously; the
// fairness across competing Receivers is whatever Go's channel runtime
// guarantees (FIFO across the goroutines blocked on the same channel).
type pollable[T any] struct {
	buf       chan Message[T]
	closed    atomic.Bool
	closeOnce sync.Once
}

// NewPollableChannel constructs a PollableChannel[T] with the given
// options. The buffer capacity is configured via WithBufferSize
// (default defaultBufferSize). The channel is immediately usable —
// there is no Start step.
func NewPollableChannel[T any](opts ...Option) PollableChannel[T] {
	options := NewOptions(opts...)

	return &pollable[T]{
		buf: make(chan Message[T], options.bufferSize),
	}
}

// Send enqueues msg into the internal buffer. When the buffer has
// capacity Send returns nil immediately; otherwise Send blocks until
// a slot opens, ctx expires, or Close runs. Returns
// ErrSend(ErrContextNil) on nil ctx, ErrSend(ErrClosed) after Close,
// and ErrSend(ErrTimeout, ctx.Err) when ctx cancels while blocked.
func (c *pollable[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(c, "PollableChannel is nil")

	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	if c.closed.Load() {
		return ErrSend(ErrClosed)
	}

	// Fast path: try to enqueue without blocking. If the buffer has
	// capacity we avoid the select + ctx.Done channel allocation cost.
	select {
	case c.buf <- msg:
		return nil
	default:
	}

	select {
	case c.buf <- msg:
		return nil
	case <-ctx.Done():
		return ErrSend(ErrTimeout, ctx.Err())
	}
}

// Receive blocks until a message is available, ctx expires, or the
// channel is closed AND drained. After Close, Receive keeps yielding
// buffered messages until the buffer is empty; only then does it
// return ErrReceive(ErrChannelClosed). This drain-then-close protocol
// matches the documented Spring PollableChannel semantics.
func (c *pollable[T]) Receive(ctx context.Context) (Message[T], error) {
	cassert.NotNil(c, "PollableChannel is nil")

	var zero Message[T]

	if ctx == nil {
		return zero, ErrReceive(ErrContextNil)
	}

	// Fast path: if a message is already buffered, take it without
	// touching ctx.Done. This also matters for the post-Close drain —
	// a closed Go channel still delivers buffered values via the same
	// receive op.
	select {
	case msg, ok := <-c.buf:
		if !ok {
			return zero, ErrReceive(ErrChannelClosed)
		}

		return msg, nil
	default:
	}

	select {
	case msg, ok := <-c.buf:
		if !ok {
			return zero, ErrReceive(ErrChannelClosed)
		}

		return msg, nil
	case <-ctx.Done():
		return zero, ErrReceive(ErrTimeout, ctx.Err())
	}
}

// Close marks the channel as closed. Subsequent Send calls return
// ErrSend(ErrClosed); pending and buffered messages remain receivable
// until the buffer is drained, at which point Receive returns
// ErrReceive(ErrChannelClosed). Close is idempotent.
func (c *pollable[T]) Close() error {
	cassert.NotNil(c, "PollableChannel is nil")

	c.closeOnce.Do(func() {
		c.closed.Store(true)
		close(c.buf)
	})

	return nil
}
