package breaker

import (
	"errors"

	"github.com/sony/gobreaker"
)

// settingsFor builds a gobreaker.Settings from the configured options.
// The ReadyToTrip predicate triggers when consecutive failures reach the
// configured threshold. The OnStateChange callback bridges gobreaker's
// state type into the package's State enum so the caller's hook sees
// domain types.
func settingsFor(opts *Options) gobreaker.Settings {
	threshold := opts.consecutiveFailures
	hook := opts.onStateChange

	return gobreaker.Settings{
		MaxRequests: opts.maxRequests,
		Interval:    opts.interval,
		Timeout:     opts.timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= threshold
		},
		OnStateChange: func(_ string, from gobreaker.State, to gobreaker.State) {
			hook(fromGobreakerState(from), fromGobreakerState(to))
		},
	}
}

// fromGobreakerState converts a gobreaker.State into the package State.
// Unknown values fall back to StateClosed so a future gobreaker addition
// cannot panic this package.
func fromGobreakerState(s gobreaker.State) State {
	switch s {
	case gobreaker.StateClosed:
		return StateClosed
	case gobreaker.StateHalfOpen:
		return StateHalfOpen
	case gobreaker.StateOpen:
		return StateOpen
	default:
		return StateClosed
	}
}

// translateBreakerError maps gobreaker sentinel errors to this package's
// sentinels, leaving any other error untouched so the caller's failure is
// preserved.
func translateBreakerError(err error) error {
	if errors.Is(err, gobreaker.ErrOpenState) {
		return ErrBreakerOpen
	}
	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		return ErrBreakerTooManyRequests
	}
	return err
}
