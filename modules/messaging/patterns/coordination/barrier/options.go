package barrier

import (
	"time"

	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring barrier Options.
// Barrier has no T-typed options, so Option is non-generic — matching
// bridge/filter/history and diverging from router/Option[T] is
// intentional.
type Option func(opts *Options)

// Options holds the configuration for a Barrier.
type Options struct {
	groupTimeout   time.Duration
	maxGroups      int
	sweepInterval  time.Duration
	errorHandler   messaging.ErrorHandler
	dropHandler    DropHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. groupTimeout has no default — it must be set via
// WithGroupTimeout because unbounded accumulation is a memory leak.
// The default maxGroups is DefaultMaxGroups; the default
// sweepInterval is DefaultSweepInterval; the default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// DropHandler is nil (silent unless wired).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		groupTimeout:  0,
		maxGroups:     DefaultMaxGroups,
		sweepInterval: DefaultSweepInterval,
		errorHandler:  messaging.DefaultErrorHandler,
		dropHandler:   nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithGroupTimeout sets the per-group timeout. A group whose quorum
// is not reached within this duration is dropped by the sweeper
// (every accumulated message fires WithDropHandler). The constructor
// REQUIRES a positive value to bound memory; without it a misbehaving
// upstream that never sends the full quorum would leak memory forever.
// Non-positive values are ignored.
func WithGroupTimeout(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.groupTimeout = d
		}
	}
}

// WithMaxGroups caps the number of distinct CorrelationIDs the
// Barrier tracks at any one time. New correlations arriving while the
// in-flight group count is already at the cap are dropped via
// WithDropHandler. Non-positive values are ignored.
func WithMaxGroups(n int) Option {
	return func(opts *Options) {
		if n > 0 {
			opts.maxGroups = n
		}
	}
}

// WithSweepInterval sets the cadence at which the sweeper inspects
// groups for timeout eviction. Smaller intervals reduce time-to-
// eviction but increase CPU; the default (DefaultSweepInterval) is a
// reasonable balance. Non-positive values are ignored.
func WithSweepInterval(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.sweepInterval = d
		}
	}
}

// WithErrorHandler installs an observability hook fired once per
// forward Send failure during release. The default (when
// WithErrorHandler is not passed) is messaging.DefaultErrorHandler,
// which logs each failure via common/log so consumers that forget to
// wire observability still see forward failures. Pass
// messaging.SilentErrorHandler to opt out, or any custom hook to
// redirect. Nil values are ignored (the previously installed handler
// is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (missing correlation, MaxGroups cap, quorum
// timeout, Stop drain). The default (when WithDropHandler is not
// passed) is nil — intentional drops are silent. Wire this to ship
// throughput metrics or audit trails. Nil arguments are ignored (the
// previously installed handler is preserved).
func WithDropHandler(handler DropHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}
