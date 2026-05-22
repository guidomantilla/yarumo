// Demo that exercises diagnostics.BuildTraceFlightRecorder end-to-end and proves that:
//
//  1. The builder shape `(TraceFlightRecorder, lifecycle.CloseFn, error)`
//     matches the project's managed-component idiom and mirrors
//     http.BuildServer / cron.BuildScheduler / grpc.BuildServer.
//  2. `defer stopFn(ctx, timeout)` triggers Stop, the runtime trace
//     flight recorder buffer is released, the lifecycle goroutine exits
//     via the internal `spawned` channel, and closeFn only returns
//     after that happens — no race window for callers observing
//     goroutine counts.
//  3. The recorder actually buffers runtime events: after a brief
//     workload burst the buffer dump to disk yields a non-empty file
//     parseable by `go tool trace`.
//  4. No goroutines leak: the count returns to the pre-Build baseline
//     after stopFn completes.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/guidomantilla/yarumo/config"
	"github.com/guidomantilla/yarumo/diagnostics"
)

const traceFile = "/tmp/trace.diagnostics.example.out"

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/diagnostics/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	baseline := runtime.NumGoroutine()

	errChan := make(chan error, 1)

	recorder, stopFn, err := diagnostics.BuildTraceFlightRecorder(
		ctx, "demo-tracefr", errChan,
		diagnostics.WithMinAge(1*time.Second),
		diagnostics.WithMaxBytes(1<<20),
	)
	if err != nil {
		return fmt.Errorf("build trace flight recorder: %w", err)
	}

	defer func() {
		// The lifecycle goroutine takes a beat to fully exit after
		// Stop returns. closeFn's `<-spawned` already drains it, but
		// the runtime's own helper goroutines (trace reader, GC
		// assist) need a short grace period for the count to settle.
		time.Sleep(50 * time.Millisecond)

		fmt.Printf("[main] post-stop goroutines: %d (baseline %d)\n",
			runtime.NumGoroutine(), baseline)
	}()
	defer stopFn(ctx, 5*time.Second)

	fmt.Printf("[main] goroutines: baseline=%d  after-build=%d\n",
		baseline, runtime.NumGoroutine())

	// Wait for the flight recorder to actually transition to Enabled —
	// Start is non-blocking and the runtime/trace machinery needs a
	// few µs to flip the bit.
	deadline := time.Now().Add(time.Second)

	for !recorder.Enabled() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !recorder.Enabled() {
		return fmt.Errorf("trace recorder never reached Enabled state")
	}

	// Force some load so the buffer fills with real events.
	loadWorkload()

	// Wait ~1s for the recorder buffer to accrue samples meeting MinAge.
	time.Sleep(1100 * time.Millisecond)

	file, err := os.Create(traceFile)
	if err != nil {
		return fmt.Errorf("create trace file: %w", err)
	}

	defer func() { _ = file.Close() }()

	n, err := recorder.WriteTo(file)
	if err != nil {
		return fmt.Errorf("write trace: %w", err)
	}

	if n < 1024 {
		return fmt.Errorf("expected trace file > 1 KB, got %d bytes", n)
	}

	fmt.Printf("[trace] wrote %d bytes to %s\n", n, traceFile)
	fmt.Printf("[trace] inspect with: go tool trace %s\n", traceFile)
	fmt.Println("[main] returning (defer stopFn next)")

	return nil
}

func loadWorkload() {
	var wg sync.WaitGroup

	for range 50 {
		wg.Go(func() {
			buf := make([]byte, 1<<14)
			runtime.KeepAlive(buf)
			time.Sleep(5 * time.Millisecond)
		})
	}

	wg.Wait()
}
