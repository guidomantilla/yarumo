package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

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
