package splitter

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// splitter is the Splitter implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled
// in Stop) and emits one Message[U] per item returned by the
// configured SplitFn for each received Message[T].
type splitter[T, U any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[U]
	split        SplitFn[T, U]
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewSplitter constructs a Splitter that subscribes to src and emits
// one Message[U] per item returned by split for each incoming
// Message[T]. The splitter is not running on return; call
// lifecycle.Build (or Start directly) to register the subscription.
//
// name is used in lifecycle logs and must be non-empty. src, dst and
// split are mandatory. Optional behaviors:
//
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for SplitFn errors/panics and forward failures.
//   - WithDropHandler installs an optional hook for observing
//     intentional drops (empty-slice returns); nil by default (silent
//     drop).
//
// Each emitted child carries Headers.CorrelationID from the source
// message, Headers.CausationID = source MessageID,
// Headers.MessageID = `<source MessageID>-<index>`,
// Headers.SequenceNumber = 0-based index, and Headers.SequenceSize =
// len(slice). All other Headers fields are preserved verbatim.
func NewSplitter[T, U any](name string, src messaging.Channel[T], dst messaging.Channel[U], split SplitFn[T, U], opts ...Option[U]) Splitter[T, U] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")
	cassert.NotNil(split, "split function is nil")

	options := NewOptions(opts...)

	return &splitter[T, U]{
		name:         name,
		src:          src,
		dst:          dst,
		split:        split,
		errorHandler: options.errorHandler,
		dropHandler:  options.dropHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the splitter's identity used in lifecycle logs.
func (s *splitter[T, U]) Name() string {
	cassert.NotNil(s, "splitter is nil")

	return s.name
}

// Start registers the splitting handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (s *splitter[T, U]) Start(_ context.Context) error {
	cassert.NotNil(s, "splitter is nil")

	var startErr error

	s.startOnce.Do(func() {
		cancel, err := s.src.Subscribe(s.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		s.mu.Lock()
		s.cancel = cancel
		s.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (s *splitter[T, U]) Stop(ctx context.Context) error {
	cassert.NotNil(s, "splitter is nil")

	s.stopOnce.Do(func() {
		s.mu.Lock()
		cancel := s.cancel
		s.cancel = nil
		s.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		s.doneOnce.Do(func() { close(s.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (s *splitter[T, U]) Done() <-chan struct{} {
	cassert.NotNil(s, "splitter is nil")

	return s.done
}

// handle is the Handler[T] subscribed on the source channel. It runs
// the SplitFn under panic recovery, then emits one Message[U] per
// returned item to the destination. Empty slices route to the
// DropHandler. The function itself always returns nil so splitter
// concerns never propagate to the source channel's Send caller.
func (s *splitter[T, U]) handle(ctx context.Context, msg messaging.Message[T]) error {
	items, err := s.splitWithRecover(ctx, msg)
	if err != nil {
		s.reportError(ctx, msg, err)

		return nil
	}

	if len(items) == 0 {
		s.reportDrop(ctx, msg)

		return nil
	}

	s.emitChildren(ctx, msg, items)

	return nil
}

// splitWithRecover invokes the user-supplied SplitFn under panic
// recovery. Panics become ErrSplitter(ErrSplitterPanic, ...) errors;
// normal errors become ErrSplitter(ErrSplitFailed, err).
func (s *splitter[T, U]) splitWithRecover(ctx context.Context, msg messaging.Message[T]) (items []U, err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		items = nil
		err = ErrSplitter(ErrSplitterPanic, fmt.Errorf("%v", rec))
	}()

	items, err = s.split(ctx, msg)
	if err != nil {
		return nil, ErrSplitter(ErrSplitFailed, err)
	}

	return items, nil
}

// emitChildren builds and sends one Message[U] per item, populating the
// sequence headers and lineage fields. A failed Send is reported but
// does not abort the batch — remaining children are still emitted.
func (s *splitter[T, U]) emitChildren(ctx context.Context, original messaging.Message[T], items []U) {
	size := len(items)

	for idx, item := range items {
		child := s.buildChild(original, item, idx, size)

		err := s.dst.Send(ctx, child)
		if err != nil {
			s.reportError(ctx, original, ErrSplitter(ErrForwardFailed, err))
		}
	}
}

// buildChild constructs the child Message[U] for index idx of size
// total items. Headers lineage policy is documented in the package
// doc.
func (s *splitter[T, U]) buildChild(original messaging.Message[T], item U, idx, size int) messaging.Message[U] {
	headers := original.Headers
	headers.MessageID = childMessageID(original.Headers.MessageID, idx)
	headers.CausationID = original.Headers.MessageID
	headers.SequenceNumber = idx
	headers.SequenceSize = size

	return messaging.Message[U]{
		Payload: item,
		Headers: headers,
	}
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (s *splitter[T, U]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if s.errorHandler == nil {
		return
	}

	s.errorHandler(ctx, msg, err)
}

// reportDrop forwards msg to the configured DropHandler. DropHandler
// is nil by default (silent drops); the guard skips invocation in that
// case.
func (s *splitter[T, U]) reportDrop(ctx context.Context, msg messaging.Message[T]) {
	if s.dropHandler == nil {
		return
	}

	s.dropHandler(ctx, msg)
}

// childMessageID builds the MessageID for a child at index idx given
// the source MessageID. When the source has no MessageID, the child
// carries just the index — uncommon but well-defined.
func childMessageID(sourceID string, idx int) string {
	if sourceID == "" {
		return strconv.Itoa(idx)
	}

	return sourceID + "-" + strconv.Itoa(idx)
}
