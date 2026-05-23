// Demo that exercises every public entry point of modules/diagnostics:
//
//  1. CaptureCPUProfile         — one-shot CPU profile to disk (blocks for duration).
//  2. CaptureHeapProfile        — one-shot heap snapshot to disk.
//  3. CaptureGoroutineProfile   — one-shot goroutine snapshot to disk.
//  4. CaptureBlockProfile       — one-shot block profile to disk (requires
//                                 SetBlockProfileRate > 0 first).
//  5. NewPprofHandler           — HTTP handler exposing /debug/pprof/* on a
//                                 httptest.Server, with a sanity fetch of the
//                                 index page.
//  6. BuildBlockProfiling       — lifecycle.Component that owns the
//                                 SetBlockProfileRate toggle: Start enables
//                                 sampling, Stop resets to zero.
//  7. BuildTraceFlightRecorder  — lifecycle.Component wrapping the Go 1.25+
//                                 runtime/trace.FlightRecorder: continuous
//                                 buffer, dumped on demand.
//
// Each demo runs sequentially with a clear header. Each captured profile is
// written under /tmp/diagnostics.example.<name>.out for offline inspection
// with `go tool pprof <file>` or `go tool trace <file>`.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/guidomantilla/yarumo/config"
	"github.com/guidomantilla/yarumo/diagnostics"
)

const outputDir = "/tmp"

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
	fmt.Printf("[main] baseline goroutines: %d\n\n", baseline)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"CaptureCPUProfile (one-shot)", demoCaptureCPUProfile},
		{"CaptureHeapProfile (one-shot)", demoCaptureHeapProfile},
		{"CaptureGoroutineProfile (one-shot)", demoCaptureGoroutineProfile},
		{"CaptureBlockProfile (one-shot, requires SetBlockProfileRate)", demoCaptureBlockProfile},
		{"NewPprofHandler (HTTP exposition)", demoPprofHandler},
		{"BuildBlockProfiling (lifecycle.Component)", demoBuildBlockProfiling},
		{"BuildTraceFlightRecorder (lifecycle.Component)", demoBuildTraceFlightRecorder},
	}

	for _, d := range demos {
		fmt.Printf("=== Demo: %s ===\n", d.title)

		err := d.fn(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", d.title, err)
		}

		fmt.Println()
	}

	// Brief grace period for any lifecycle helpers to fully unwind.
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("[main] post-demo goroutines: %d (baseline %d)\n",
		runtime.NumGoroutine(), baseline)

	return nil
}

// demoCaptureCPUProfile captures CPU samples for a short window and writes
// the profile to disk. The function blocks for the configured duration.
func demoCaptureCPUProfile(ctx context.Context) error {
	path := outputDir + "/diagnostics.example.cpu.pprof"

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create cpu profile file: %w", err)
	}

	defer func() { _ = file.Close() }()

	// Run a workload concurrently so the CPU profiler has work to sample.
	done := make(chan struct{})
	go func() {
		defer close(done)
		cpuWorkload(200 * time.Millisecond)
	}()

	err = diagnostics.CaptureCPUProfile(ctx, file, 200*time.Millisecond)
	if err != nil {
		return fmt.Errorf("CaptureCPUProfile: %w", err)
	}

	<-done

	size, _ := fileSize(path)
	fmt.Printf("[cpu] wrote %d bytes to %s\n", size, path)
	fmt.Printf("[cpu] inspect with: go tool pprof %s\n", path)

	return nil
}

// demoCaptureHeapProfile writes a heap snapshot to disk. The runtime
// always samples allocations, so this has near-zero overhead beyond the
// cost of writing the snapshot.
func demoCaptureHeapProfile(_ context.Context) error {
	path := outputDir + "/diagnostics.example.heap.pprof"

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create heap profile file: %w", err)
	}

	defer func() { _ = file.Close() }()

	err = diagnostics.CaptureHeapProfile(file)
	if err != nil {
		return fmt.Errorf("CaptureHeapProfile: %w", err)
	}

	size, _ := fileSize(path)
	fmt.Printf("[heap] wrote %d bytes to %s\n", size, path)
	fmt.Printf("[heap] inspect with: go tool pprof %s\n", path)

	return nil
}

// demoCaptureGoroutineProfile writes a goroutine snapshot to disk. Capture
// cost scales with the number of live goroutines at the moment of capture.
func demoCaptureGoroutineProfile(_ context.Context) error {
	path := outputDir + "/diagnostics.example.goroutine.pprof"

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create goroutine profile file: %w", err)
	}

	defer func() { _ = file.Close() }()

	err = diagnostics.CaptureGoroutineProfile(file)
	if err != nil {
		return fmt.Errorf("CaptureGoroutineProfile: %w", err)
	}

	size, _ := fileSize(path)
	fmt.Printf("[goroutine] wrote %d bytes to %s (snapshot of %d goroutines)\n",
		size, path, runtime.NumGoroutine())
	fmt.Printf("[goroutine] inspect with: go tool pprof %s\n", path)

	return nil
}

// demoCaptureBlockProfile enables block-profile sampling, runs a workload
// that blocks on synchronisation primitives, captures the profile, and
// disables sampling. CaptureBlockProfile by itself yields an empty profile
// unless SetBlockProfileRate has been called with a non-zero rate.
func demoCaptureBlockProfile(_ context.Context) error {
	path := outputDir + "/diagnostics.example.block.pprof"

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create block profile file: %w", err)
	}

	defer func() { _ = file.Close() }()

	// Enable block-profile sampling for the duration of the demo. Rate 1
	// records every blocking event; production usage prefers higher rates
	// (e.g. 10000) to cap overhead.
	runtime.SetBlockProfileRate(1)

	defer runtime.SetBlockProfileRate(0)

	blockWorkload(50 * time.Millisecond)

	err = diagnostics.CaptureBlockProfile(file)
	if err != nil {
		return fmt.Errorf("CaptureBlockProfile: %w", err)
	}

	size, _ := fileSize(path)
	fmt.Printf("[block] wrote %d bytes to %s\n", size, path)
	fmt.Printf("[block] inspect with: go tool pprof %s\n", path)

	return nil
}

// demoPprofHandler mounts NewPprofHandler on a httptest.Server and fetches
// /debug/pprof/ to confirm the index page is served. In a real service the
// handler is mounted on the application router (e.g. on the management HTTP
// server bound to a non-public port).
func demoPprofHandler(_ context.Context) error {
	handler := diagnostics.NewPprofHandler()
	server := httptest.NewServer(handler)

	defer server.Close()

	url := server.URL + "/debug/pprof/"

	//nolint:noctx
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200, got %d", resp.StatusCode)
	}

	fmt.Printf("[pprof-handler] served on %s\n", server.URL)
	fmt.Printf("[pprof-handler] GET /debug/pprof/ -> %d, %d bytes index\n",
		resp.StatusCode, len(body))
	fmt.Printf("[pprof-handler] in real apps, mount it on a management router:\n")
	fmt.Printf("                mux.Handle(\"/debug/pprof/\", diagnostics.NewPprofHandler())\n")

	return nil
}

// demoBuildBlockProfiling exercises the lifecycle component that owns
// runtime.SetBlockProfileRate: Start enables sampling, Stop resets to
// zero. The component shape matches http.BuildServer, cron.BuildScheduler,
// grpc.BuildServer, and BuildTraceFlightRecorder.
func demoBuildBlockProfiling(ctx context.Context) error {
	errChan := make(chan error, 1)

	sampler, stopFn, err := diagnostics.BuildBlockProfiling(
		ctx, "demo-blockprof", errChan,
		diagnostics.WithBlockProfileRate(1),
	)
	if err != nil {
		return fmt.Errorf("BuildBlockProfiling: %w", err)
	}

	defer stopFn(ctx, 2*time.Second)

	// Wait briefly for Start to flip runtime.SetBlockProfileRate.
	time.Sleep(20 * time.Millisecond)

	blockWorkload(50 * time.Millisecond)

	path := outputDir + "/diagnostics.example.blockprof-component.pprof"

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create block profile file: %w", err)
	}

	defer func() { _ = file.Close() }()

	err = diagnostics.CaptureBlockProfile(file)
	if err != nil {
		return fmt.Errorf("CaptureBlockProfile (component-managed rate): %w", err)
	}

	size, _ := fileSize(path)
	fmt.Printf("[blockprof-component] rate=%d, wrote %d bytes to %s\n",
		sampler.Rate(), size, path)
	fmt.Printf("[blockprof-component] Stop will reset rate to 0 via stopFn\n")

	return nil
}

// demoBuildTraceFlightRecorder exercises the lifecycle component wrapping
// runtime/trace.FlightRecorder. The recorder continuously buffers runtime
// events; WriteTo dumps the current buffer on demand. Only one flight
// recorder may be active per process at a time.
func demoBuildTraceFlightRecorder(ctx context.Context) error {
	errChan := make(chan error, 1)

	recorder, stopFn, err := diagnostics.BuildTraceFlightRecorder(
		ctx, "demo-tracefr", errChan,
		diagnostics.WithMinAge(1*time.Second),
		diagnostics.WithMaxBytes(1<<20),
	)
	if err != nil {
		return fmt.Errorf("BuildTraceFlightRecorder: %w", err)
	}

	defer stopFn(ctx, 5*time.Second)

	// Wait for Start to flip the recorder into Enabled — runtime/trace
	// needs a few µs to flip the bit after Start returns.
	deadline := time.Now().Add(time.Second)
	for !recorder.Enabled() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !recorder.Enabled() {
		return fmt.Errorf("trace recorder never reached Enabled state")
	}

	// Force some load so the buffer fills with real events.
	loadWorkload()

	// Wait ~1s so the buffer accrues samples meeting MinAge.
	time.Sleep(1100 * time.Millisecond)

	path := outputDir + "/diagnostics.example.trace.out"

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create trace file: %w", err)
	}

	defer func() { _ = file.Close() }()

	n, err := recorder.WriteTo(file)
	if err != nil {
		return fmt.Errorf("WriteTo: %w", err)
	}

	if n < 1024 {
		return fmt.Errorf("expected trace file > 1 KB, got %d bytes", n)
	}

	fmt.Printf("[tracefr] wrote %d bytes to %s\n", n, path)
	fmt.Printf("[tracefr] inspect with: go tool trace %s\n", path)

	return nil
}

// cpuWorkload spins the CPU long enough for the profiler to record
// non-trivial samples. The work itself is meaningless; what matters is that
// the goroutine stays on-CPU.
func cpuWorkload(d time.Duration) {
	deadline := time.Now().Add(d)
	x := uint64(1)

	for time.Now().Before(deadline) {
		for range 1000 {
			x = x*1103515245 + 12345
		}
	}

	runtime.KeepAlive(x)
}

// blockWorkload generates goroutine contention so the block profile has
// events to sample. Goroutines contend on a single mutex held in short
// bursts.
func blockWorkload(d time.Duration) {
	var mu sync.Mutex

	deadline := time.Now().Add(d)

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for time.Now().Before(deadline) {
				mu.Lock()
				time.Sleep(time.Microsecond)
				mu.Unlock()
			}
		})
	}

	wg.Wait()
}

// loadWorkload spawns many short-lived goroutines so the trace recorder's
// buffer fills with goroutine create/run/end events.
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

// fileSize returns the size in bytes of the file at path, or zero if Stat
// fails. Used for human-readable output only.
func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}
