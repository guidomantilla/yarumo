package cron

import (
	"context"
	"time"

	cron "github.com/robfig/cron/v3"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	clog "github.com/guidomantilla/yarumo/log"
)

// BuildScheduler creates a managed cron Scheduler, starts it in a background
// goroutine, and returns a CloseFn for graceful shutdown.
//
// Startup errors are logged and forwarded to errChan (non-blocking). The
// returned Scheduler can be used to register jobs via AddFunc / AddJob /
// Schedule. The returned CloseFn must be called by the caller to release
// resources; it bounds the shutdown by the given timeout and blocks until
// the background goroutine has exited.
func BuildScheduler(ctx context.Context, name string, errChan lifecycle.ErrChan, options ...cron.Option) (Scheduler, lifecycle.CloseFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	component := NewScheduler(name, options...)

	spawned := make(chan struct{})

	closeFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		err := lifecycle.Stop(ctx, component, timeout)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}

		<-spawned
	}

	go func() {
		defer close(spawned)

		err := lifecycle.Start(ctx, component, errChan)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
		}
	}()

	return component, closeFn, nil
}
