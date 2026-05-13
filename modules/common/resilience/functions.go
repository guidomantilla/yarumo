package resilience

import (
	"context"
	"errors"

	"github.com/sony/gobreaker"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// validateName checks that the registry key is non-empty. The check uses an
// explicit nil/empty guard rather than cassert because the name is a
// caller-supplied runtime value rather than a struct invariant.
func validateName(name string) error {
	if name == "" {
		return cerrs.Wrap(ErrRegistryNameEmpty)
	}

	return nil
}

// validateExecute returns a non-nil sentinel error when the inputs to
// CircuitBreaker.Execute are invalid (nil context, nil fn).
func validateExecute(ctx context.Context, fn func() (any, error)) error {
	if ctx == nil {
		return cerrs.Wrap(ErrContextNil)
	}

	if fn == nil {
		return cerrs.Wrap(ErrCircuitBreakerExecuteFnNil)
	}

	return nil
}

// validateWait returns a non-nil sentinel error when the inputs to
// RateLimiter.Wait are invalid (nil context).
func validateWait(ctx context.Context) error {
	if ctx == nil {
		return cerrs.Wrap(ErrContextNil)
	}

	return nil
}

// translateBreakerError converts a gobreaker sentinel error into the
// resilience package sentinel. Non-sentinel errors are returned unchanged so
// the underlying call failure is preserved for the caller.
func translateBreakerError(err error) error {
	if errors.Is(err, gobreaker.ErrOpenState) {
		return ErrCircuitBreakerOpen
	}

	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		return ErrCircuitBreakerTooManyRequests
	}

	return err
}

// fromGobreakerState converts a gobreaker.State into the package State.
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

// settingsFor builds a gobreaker.Settings from the options.
func settingsFor(name string, opts *Options) gobreaker.Settings {
	maxFailures := opts.cbConsecutiveFailures

	return gobreaker.Settings{
		Name:        name,
		MaxRequests: opts.cbMaxRequests,
		Interval:    opts.cbInterval,
		Timeout:     opts.cbTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= maxFailures
		},
	}
}
