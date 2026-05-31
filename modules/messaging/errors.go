package messaging

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// MessagingType is the error domain identifier for messaging operations.
const MessagingType = "messaging"

var (
	_ error = (*Error)(nil)

	_ ErrSendFn      = ErrSend
	_ ErrSubscribeFn = ErrSubscribe
	_ ErrReceiveFn   = ErrReceive
)

// ErrSendFn is the function type for ErrSend.
type ErrSendFn func(causes ...error) error

// ErrSubscribeFn is the function type for ErrSubscribe.
type ErrSubscribeFn func(causes ...error) error

// ErrReceiveFn is the function type for ErrReceive.
type ErrReceiveFn func(causes ...error) error

// Sentinel errors for messaging operations.
var (
	// ErrSendFailed indicates that a Send operation failed.
	ErrSendFailed = errors.New("send failed")
	// ErrSubscribeFailed indicates that a Subscribe operation failed.
	ErrSubscribeFailed = errors.New("subscribe failed")
	// ErrClosed indicates that the channel has been closed and no
	// longer accepts Send or Subscribe.
	ErrClosed = errors.New("channel closed")
	// ErrHandlerNil indicates that a nil handler was passed to
	// Subscribe.
	ErrHandlerNil = errors.New("handler is nil")
	// ErrContextNil indicates that a nil context was passed to Send.
	ErrContextNil = errors.New("context is nil")
	// ErrTimeout indicates that an operation timed out (e.g. enqueue
	// blocked past the configured deadline).
	ErrTimeout = errors.New("operation timed out")
	// ErrDrainTimeout indicates that the queue did not finish draining
	// pending messages before Stop's context deadline expired.
	ErrDrainTimeout = errors.New("drain timeout")
	// ErrHandlerPanic indicates that a pipeline handler panicked
	// during dispatch. The recovered value is embedded in the
	// resulting StepResult.Err via this sentinel.
	ErrHandlerPanic = errors.New("handler panicked")
	// ErrChainFailed indicates that a PipelineChannel send aborted
	// because at least one step returned an error. The full step trace
	// is available on the returned *ChainError.
	ErrChainFailed = errors.New("pipeline chain failed")
	// ErrNoSubscribers indicates that a QueueChannel worker pulled a
	// message from the inbound buffer but no subscribers were
	// registered to receive it. The message is dropped and routed
	// through the ErrorHandler hook for visibility.
	ErrNoSubscribers = errors.New("no subscribers registered")
	// ErrDropped indicates a message was intentionally discarded by a
	// sink channel (NullChannel) or an overflow policy. Surfaced
	// through the ErrorHandler hook so test/observability paths see
	// the drop without it being a hard error to the Send caller.
	ErrDropped = errors.New("message dropped")
	// ErrOverflow indicates a buffer overflow under DropNewest or
	// DropOldest policy. Joined with ErrDropped in the hook payload so
	// errors.Is(err, ErrDropped) also matches; use
	// errors.Is(err, ErrOverflow) to distinguish overflow drops from
	// NullChannel drops.
	ErrOverflow = errors.New("buffer overflow")
	// ErrBufferFull indicates Send rejected the message because the
	// buffer is at capacity and the OverflowReject policy is active.
	// Returned from Send (NOT routed through the hook) so callers can
	// implement their own retry / backoff / shedding logic.
	ErrBufferFull = errors.New("buffer full")
	// ErrReceiveFailed indicates that a Receive operation failed.
	// Wrapped by ErrReceive(...) factory on PollableChannel.Receive
	// returns.
	ErrReceiveFailed = errors.New("receive failed")
	// ErrChannelClosed indicates that a PollableChannel.Receive was
	// invoked (or unblocked) after the channel was closed and the
	// internal buffer is drained. Distinct from ErrClosed (which
	// covers Send/Subscribe on a stopped Channel[T]) to make the
	// pollable consumer's drain-then-exit loop unambiguous.
	ErrChannelClosed = errors.New("channel closed for receive")
)

// StepStatus classifies the outcome of a single pipeline step in a
// ChainError trace.
type StepStatus string

const (
	// StepStatusOK indicates the step completed without error.
	StepStatusOK StepStatus = "ok"
	// StepStatusError indicates the step returned a non-nil error.
	StepStatusError StepStatus = "error"
	// StepStatusPanic indicates the step panicked during execution.
	StepStatusPanic StepStatus = "panic"
	// StepStatusSkipped indicates the step was not executed because a
	// previous step in the chain failed (fail-fast).
	StepStatusSkipped StepStatus = "skipped"
)

// StepResult records the outcome of a single handler invocation
// within a PipelineChannel.Send call.
type StepResult struct {
	// Index is the 0-based position of the step in the chain, in
	// Subscribe order.
	Index int
	// Status reports the outcome (ok / error / panic / skipped).
	Status StepStatus
	// Err is the non-nil error returned by the handler when Status is
	// StepStatusError, the wrapped recovered value when Status is
	// StepStatusPanic, or nil otherwise.
	Err error
	// Duration is the wall-clock time the handler took to execute.
	// Zero for skipped steps.
	Duration time.Duration
}

// ChainError is the error returned by PipelineChannel.Send when at
// least one step fails. It carries a per-step trace (ok / error /
// panic / skipped) so callers can render which steps ran, which one
// broke, and which never executed. ChainError.Unwrap returns the
// failing step's underlying error so errors.Is/As keep working.
type ChainError struct {
	// Steps holds the outcome of every step in the chain, in order.
	// Failed step's index is also exposed via Failed.
	Steps []StepResult
	// Failed is the index of the step that aborted the chain, or -1
	// when every step completed successfully (in which case Send
	// returns nil instead of *ChainError).
	Failed int
}

// Error renders the chain trace as a multi-line string suitable for
// logs.
func (e *ChainError) Error() string {
	cassert.NotNil(e, "ChainError is nil")

	var b strings.Builder

	b.WriteString("pipeline chain failed at step ")
	b.WriteString(strconv.Itoa(e.Failed))
	b.WriteString(":\n")

	for _, step := range e.Steps {
		b.WriteString("  [")
		b.WriteString(strconv.Itoa(step.Index))
		b.WriteString("] ")
		b.WriteString(string(step.Status))

		for i := len(step.Status); i < 8; i++ {
			b.WriteByte(' ')
		}

		b.WriteString(" (")
		b.WriteString(step.Duration.String())
		b.WriteByte(')')

		if step.Err != nil {
			b.WriteString(": ")
			b.WriteString(step.Err.Error())
		}

		b.WriteByte('\n')
	}

	return strings.TrimRight(b.String(), "\n")
}

// Unwrap returns the failing step's underlying error so errors.Is and
// errors.As can match against the original sentinel.
func (e *ChainError) Unwrap() error {
	cassert.NotNil(e, "ChainError is nil")

	if e.Failed < 0 || e.Failed >= len(e.Steps) {
		return nil
	}

	return e.Steps[e.Failed].Err
}

// Error is the domain error type for messaging operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("messaging %s error: %s", e.Type, e.Err)
}

// ErrSend wraps the given causes into a domain Error for Send failures.
func ErrSend(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: MessagingType,
			Err:  errors.Join(append(causes, ErrSendFailed)...),
		},
	}
}

// ErrSubscribe wraps the given causes into a domain Error for
// Subscribe failures.
func ErrSubscribe(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: MessagingType,
			Err:  errors.Join(append(causes, ErrSubscribeFailed)...),
		},
	}
}

// ErrReceive wraps the given causes into a domain Error for Receive
// failures on a PollableChannel.
func ErrReceive(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: MessagingType,
			Err:  errors.Join(append(causes, ErrReceiveFailed)...),
		},
	}
}
