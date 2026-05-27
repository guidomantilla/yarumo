package breaker

import (
	"errors"

	"github.com/sony/gobreaker"

	cbreaker "github.com/guidomantilla/yarumo/core/common/resilience/breaker"
)

// settingsFor builds a gobreaker.Settings from the configured options.
// The ReadyToTrip predicate triggers when consecutive failures reach the
// configured threshold. The OnStateChange callback bridges gobreaker's
// state type into the contract's State enum so the caller's hook sees
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

// fromGobreakerState converts a gobreaker.State into the contract State.
// Unknown values fall back to StateClosed so a future gobreaker addition
// cannot panic this package.
func fromGobreakerState(s gobreaker.State) cbreaker.State {
	switch s {
	case gobreaker.StateClosed:
		return cbreaker.StateClosed
	case gobreaker.StateHalfOpen:
		return cbreaker.StateHalfOpen
	case gobreaker.StateOpen:
		return cbreaker.StateOpen
	default:
		return cbreaker.StateClosed
	}
}

// translateBreakerError maps gobreaker sentinel errors to the contract
// sentinels, leaving any other error untouched so the caller's failure is
// preserved.
func translateBreakerError(err error) error {
	if errors.Is(err, gobreaker.ErrOpenState) {
		return cbreaker.ErrBreakerOpen
	}
	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		return cbreaker.ErrBreakerTooManyRequests
	}
	return err
}
