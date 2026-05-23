package lifecycle

import (
	"context"
	"time"

	clog "github.com/guidomantilla/yarumo/common/log"
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

func Build(ctx context.Context, component Component, errChan ErrChan) (CloseFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", component.Name())

	spawned := make(chan struct{})

	closeFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", component.Name())
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", component.Name())

		err := Stop(ctx, component, timeout)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", component.Name(), "error", err)
		}

		<-spawned
	}

	go func() {
		defer close(spawned)

		err := Start(ctx, component, errChan)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", component.Name(), "error", err)
		}
	}()

	return closeFn, nil
}
