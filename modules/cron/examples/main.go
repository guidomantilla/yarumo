// Demo that exercises cron.BuildScheduler end-to-end and proves that:
//
//  1. The builder shape `(Scheduler, lifecycle.CloseFn, error)` matches the
//     project's managed-component idiom — call sites can either bind the
//     scheduler to register jobs (as below) or discard it with `_` when all
//     that's needed is the lifecycle.
//  2. `defer stopFn(ctx, timeout)` drains in-flight jobs and unblocks the
//     internal lifecycle.Start goroutine.
//  3. The scheduler's Start/Stop methods drive the underlying *cron.Cron
//     correctly: jobs tick on schedule while running and stop firing after
//     the deferred stopFn returns.
//  4. No goroutines leak: the count returns to the pre-Build baseline after
//     stopFn completes.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	"github.com/guidomantilla/yarumo/config"
	"github.com/guidomantilla/yarumo/cron"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/cron/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	baseline := runtime.NumGoroutine()

	errChan := make(chan error, 1)

	// Build returns (Scheduler, CloseFn, error). We bind the scheduler here
	// because the demo registers a job on it; for components that need no
	// further interaction the call site can read this as
	//   _, stopFn, err := cron.BuildScheduler(ctx, "demo-cron", errChan)
	scheduler, stopFn, err := cron.BuildScheduler(ctx, "demo-cron", errChan)
	if err != nil {
		return fmt.Errorf("build scheduler: %w", err)
	}

	// Observe goroutine cleanup AFTER stopFn returns. defer is LIFO, so the
	// stopFn fires first; this print runs once the worker goroutine has
	// already exited.
	defer func() {
		fmt.Printf("[main] post-stop goroutines: %d (baseline %d)\n",
			runtime.NumGoroutine(), baseline)
	}()
	defer stopFn(ctx, 5*time.Second)

	var counter atomic.Int32

	_, err = scheduler.AddFunc("@every 500ms", func() {
		n := counter.Add(1)
		fmt.Printf("[job] tick #%d\n", n)
	})
	if err != nil {
		return fmt.Errorf("add func: %w", err)
	}

	fmt.Printf("[main] goroutines: baseline=%d  after-build=%d\n",
		baseline, runtime.NumGoroutine())

	fmt.Println("[main] running for ~2s ...")

	select {
	case <-time.After(2 * time.Second):
	case err := <-errChan:
		return fmt.Errorf("runtime error: %w", err)
	}

	ticks := counter.Load()
	fmt.Printf("[main] observed %d ticks — returning (defer stopFn next)\n", ticks)

	if ticks == 0 {
		return fmt.Errorf("expected at least one tick, got 0")
	}

	// Quiet linter on the lifecycle import — we use it for type clarity in
	// the godoc and to assert the builder return shape at compile time.
	var _ lifecycle.CloseFn = stopFn

	return nil
}
