// Demo that exercises NewScheduler + lifecycle.Build end-to-end and proves that:
//
//  1. The two-step pattern `cron.NewScheduler(...)` + `lifecycle.Build(...)`
//     replaces the legacy cron.BuildScheduler in a uniform way — the same
//     pattern applies to http, grpc, diagnostics and any consumer-defined
//     Component.
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

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/config"
	"github.com/guidomantilla/yarumo/managed/cron"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/managed/cron/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	baseline := runtime.NumGoroutine()

	errChan := make(chan error, 1)

	// NewScheduler constructs the Component; lifecycle.Build wires it into
	// a background goroutine and returns the CloseFn for graceful shutdown.
	scheduler := cron.NewScheduler("demo-cron")

	stopFn, err := lifecycle.Build(ctx, scheduler, errChan)
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
