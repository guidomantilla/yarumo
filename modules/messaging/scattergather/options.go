package scattergather

import (
	"time"

	"github.com/guidomantilla/yarumo/messaging"
)

// defaultMaxConcurrentScatters bounds the in-flight expected-size map
// when the caller does not pass WithMaxConcurrentScatters. Sized to be
// comfortably above realistic scatter-gather workloads (dozens to
// hundreds of concurrent requests) while still catching runaway
// producers within a reasonable memory budget.
const defaultMaxConcurrentScatters = 1000

// Option is a functional option for configuring scattergather Options.
// It is generic over T to keep the door open for future T-typed
// options (e.g. custom CorrelationFn[T] mirroring the aggregator).
type Option[T any] func(opts *Options[T])

// Options holds the configuration for a ScatterGather.
type Options[T any] struct {
	groupTimeout          time.Duration
	maxConcurrentScatters int
	errorHandler          messaging.ErrorHandler
	dropHandler           DropHandler
}

// NewOptions creates a new Options[T] with sensible defaults and
// applies the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// MaxConcurrentScatters cap is defaultMaxConcurrentScatters; the
// default DropHandler is nil (silent intentional drops). No group
// timeout is set by default — WithGroupTimeout MUST be passed
// explicitly or NewScatterGather panics.
func NewOptions[T any](opts ...Option[T]) *Options[T] {
	options := &Options[T]{
		maxConcurrentScatters: defaultMaxConcurrentScatters,
		errorHandler:          messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithGroupTimeout configures the maximum lifetime of a single
// in-flight gather. A partial group that does not receive all
// expected replies within d is released by the internal Aggregator's
// sweeper, drops via WithDropHandler and frees its slot in the
// per-correlation expected-size map. This option is REQUIRED:
// constructing a ScatterGather without it is a caller bug because the
// pattern can stall forever on a worker that never replies.
// Non-positive values are ignored.
func WithGroupTimeout[T any](d time.Duration) Option[T] {
	return func(opts *Options[T]) {
		if d > 0 {
			opts.groupTimeout = d
		}
	}
}

// WithMaxConcurrentScatters caps the number of simultaneously tracked
// in-flight gathers. A request that would create the n+1-th entry is
// rejected without scattering and the ErrorHandler is invoked with
// ErrMaxScattersExceeded. The default is defaultMaxConcurrentScatters
// (1000). Non-positive values are ignored.
func WithMaxConcurrentScatters[T any](n int) Option[T] {
	return func(opts *Options[T]) {
		if n > 0 {
			opts.maxConcurrentScatters = n
		}
	}
}

// WithErrorHandler installs an observability hook fired once per real
// scatter-gather failure (missing worker key, scatter Send failed,
// AggregateFn failed, forward Send failed, MaxScatters exceeded). The
// default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via
// common/log so consumers that forget to wire observability still see
// failures. Pass messaging.SilentErrorHandler to opt out, or any
// custom hook to redirect. Nil values are ignored.
func WithErrorHandler[T any](handler messaging.ErrorHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (empty selector result, partial gather released by
// the Aggregator's timeout sweeper). The default (when
// WithDropHandler is not passed) is nil — intentional drops are
// silent. Wire this for throughput metrics or audit trails. Nil
// values are ignored.
func WithDropHandler[T any](handler DropHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}
