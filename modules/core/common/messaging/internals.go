package messaging

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cuids "github.com/guidomantilla/yarumo/core/common/uids"
)

// snapshotHandlers returns a stable copy of the handler list ordered by
// registration. It acquires mu for reading internally so callers use it
// as the canonical "read the handler list" step. The slice header for
// order must be passed by pointer so the helper observes the latest
// header after the lock is acquired (Subscribe may have appended to a
// new backing array between argument evaluation and lock acquisition);
// byID is a map (reference type), so the value copy at call time
// aliases the live map and remains correct under the lock.
func snapshotHandlers[T any](mu *sync.RWMutex, order *[]uint64, byID map[uint64]Handler[T]) []Handler[T] {
	mu.RLock()
	defer mu.RUnlock()

	list := *order
	out := make([]Handler[T], 0, len(list))

	for _, id := range list {
		out = append(out, byID[id])
	}

	return out
}

// invokeHandler runs one handler with panic recovery. Returned error is
// nil on success, the handler's error on a normal failure, or an
// ErrHandlerPanic-wrapping error on panic.
func invokeHandler[T any](ctx context.Context, msg Message[T], handler Handler[T]) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		err = fmt.Errorf("%w: %v", ErrHandlerPanic, r)
	}()

	return handler(ctx, msg)
}

// invokeStep runs one pipeline handler with panic recovery and timing.
// The work runs inside an anonymous function so the defer-recover lands
// BEFORE the outer return captures the StepResult value — keeping the
// caller signature free of a named return.
func invokeStep[T any](ctx context.Context, msg Message[T], index int, handler Handler[T]) StepResult {
	out := StepResult{Index: index, Status: StepStatusOK}
	start := time.Now()

	func() {
		defer func() {
			out.Duration = time.Since(start)

			r := recover()
			if r == nil {
				return
			}

			out.Status = StepStatusPanic
			out.Err = fmt.Errorf("%w: %v", ErrHandlerPanic, r)
		}()

		err := handler(ctx, msg)
		if err != nil {
			out.Status = StepStatusError
			out.Err = err
		}
	}()

	return out
}

// sendWithPolicy dispatches the enqueue to the per-policy helper. Each
// policy is implemented as its own small function (sendBlock /
// sendDropNewest / sendDropOldest / sendReject) so the dispatcher
// stays a thin switch and each strategy is independently testable.
// The default branch is Block, matching the historical semantics for
// any unknown policy value (already filtered by WithOverflowPolicy at
// the boundary).
func sendWithPolicy[T any](ctx context.Context, inbound chan envelope[T], msg Message[T], policy OverflowPolicy, hook ErrorHandler) error {
	env := envelope[T]{sendCtx: ctx, msg: msg}

	switch policy {
	case OverflowReject:
		return sendReject(env, inbound)
	case OverflowDropNewest:
		return sendDropNewest(ctx, env, msg, inbound, hook)
	case OverflowDropOldest:
		return sendDropOldest(ctx, env, msg, inbound, hook)
	default:
		return sendBlock(ctx, env, inbound)
	}
}

// sendReject returns ErrSend(ErrBufferFull) immediately when the
// buffer is full. No drops, no blocking, no hook invocation.
func sendReject[T any](env envelope[T], inbound chan envelope[T]) error {
	select {
	case inbound <- env:
		return nil
	default:
		return ErrSend(ErrBufferFull)
	}
}

// sendDropNewest discards the new message when the buffer is full,
// firing hook with ErrOverflow joined with ErrDropped so observers
// see the drop. Returns nil — Send did not fail, the message was
// intentionally dropped.
func sendDropNewest[T any](ctx context.Context, env envelope[T], msg Message[T], inbound chan envelope[T], hook ErrorHandler) error {
	select {
	case inbound <- env:
		return nil
	default:
		if hook != nil {
			hook(ctx, msg, errors.Join(ErrOverflow, ErrDropped))
		}

		return nil
	}
}

// sendDropOldest evicts the head of the buffer (firing hook with the
// evicted msg) and enqueues the new one. A best-effort single retry
// handles the rare race where the freed slot is taken by another
// producer; in that race the new message is dropped instead (hook
// fires for it). Returns nil — Send did not fail.
func sendDropOldest[T any](ctx context.Context, env envelope[T], msg Message[T], inbound chan envelope[T], hook ErrorHandler) error {
	select {
	case inbound <- env:
		return nil
	default:
	}

	select {
	case evicted := <-inbound:
		if hook != nil {
			hook(ctx, evicted.msg, errors.Join(ErrOverflow, ErrDropped))
		}
	default:
	}

	select {
	case inbound <- env:
		return nil
	default:
		if hook != nil {
			hook(ctx, msg, errors.Join(ErrOverflow, ErrDropped))
		}

		return nil
	}
}

// sendBlock waits for a slot to open or ctx to expire. Returns
// ErrSend(ErrTimeout, ctx.Err()) on cancellation; nil after
// successful enqueue. No drops, no hook.
func sendBlock[T any](ctx context.Context, env envelope[T], inbound chan envelope[T]) error {
	select {
	case inbound <- env:
		return nil
	case <-ctx.Done():
		return ErrSend(ErrTimeout, ctx.Err())
	}
}

// extractDLQ converts the type-erased Options.dlq field into a
// concrete Channel[DeadLetter[T]] at channel construction time.
// Returns nil when raw is nil (no DLQ configured); panics via cassert
// when raw holds a Channel parameterized by a different T than the
// channel being constructed (programmer error caught at build, not at
// first failed dispatch).
func extractDLQ[T any](raw any) Channel[DeadLetter[T]] {
	if raw == nil {
		return nil
	}

	typed, ok := raw.(Channel[DeadLetter[T]])
	cassert.True(ok, "WithDLQChannel type parameter does not match channel type T")

	return typed
}

// publishDeadLetter wraps the publish-to-DLQ side-effect used by
// Topic and Queue dispatchers when a handler returns an error. The
// Send error is intentionally swallowed — DLQ-of-DLQ is out of scope
// and the worker already reported the underlying handler error via
// the ErrorHandler hook.
func publishDeadLetter[T any](ctx context.Context, dlq Channel[DeadLetter[T]], msg Message[T], handlerErr error) {
	if dlq == nil {
		return
	}

	_ = dlq.Send(ctx, NewMessage(DeadLetter[T]{
		Original:  msg,
		LastError: handlerErr,
		FailedAt:  time.Now(),
	}, nil))
}

// generateID returns uid.Generate() or empty when uid is nil or the
// generator fails. Used by NewMessage to populate MessageID and
// CorrelationID independently so a failure on one does not abort the
// envelope construction.
func generateID(uid cuids.UID) string {
	if uid == nil {
		return ""
	}

	id, err := uid.Generate()
	if err != nil {
		return ""
	}

	return id
}
