package lifecycle

import (
	"context"
	"time"
)

// Start runs the component in the current goroutine and waits for completion.
//
// It invokes component.Start(ctx); on failure it logs the error, attempts a
// non-blocking send to errChan, and returns the error. On success it blocks
// on <-component.Done(), which works uniformly for blocking-Start (server-
// style) and non-blocking-Start (worker-style) implementations.
func Start(ctx context.Context, component Component, errChan ErrChan) error {
	err := component.Start(ctx)
	if err != nil {
		select {
		case errChan <- err:
		default:
		}

		return err
	}

	select {
	case <-component.Done():
	case <-ctx.Done():
	}

	return nil
}

// Stop signals graceful shutdown to the component bounded by the given
// timeout. It returns the error reported by component.Stop after logging.
func Stop(ctx context.Context, component Component, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := component.Stop(timeoutCtx)
	if err != nil {
		return err
	}

	return nil
}
